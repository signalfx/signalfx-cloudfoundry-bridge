package metrics

import (
    "golang.org/x/net/context"
    "crypto/tls"
    "log"
    "strconv"
    "time"

	//"github.com/davecgh/go-spew/spew"

    "github.com/cloudfoundry/noaa/consumer"
    "github.com/cloudfoundry/sonde-go/events"
    "github.com/signalfx/golib/datapoint"
)

type SignalFxFirehoseNozzle struct {
    MetricFilter
    config                *Config
    errs                  <-chan error
    messages              <-chan *events.Envelope
    authTokenFetcher      AuthTokenFetcher
    consumer              *consumer.Consumer
    client                SignalFxClient
    stop                  chan bool
    datapointBuffer       []*datapoint.Datapoint
    totalMessagesReceived int
    metadataFetcher       *AppMetadataFetcher
    deploymentMap         map[string]bool
    // Similar to the above
    metricsExcluded       map[string]bool
}

type AuthTokenFetcher interface {
    FetchAuthToken() string
}

func NewSignalFxFirehoseNozzle(config *Config,
                               tokenFetcher AuthTokenFetcher,
                               client SignalFxClient,
                               metadataFetcher *AppMetadataFetcher,
                               metricFilter *MetricFilter) *SignalFxFirehoseNozzle {

    return &SignalFxFirehoseNozzle{
        MetricFilter:     *metricFilter,
        config:           config,
        client:           client,
        errs:             make(<-chan error),
        messages:         make(<-chan *events.Envelope),
        stop:             make(chan bool),
        authTokenFetcher: tokenFetcher,
        datapointBuffer:  make([]*datapoint.Datapoint, 0, 10000),
        metadataFetcher:  metadataFetcher,
    }
}

func (o *SignalFxFirehoseNozzle) Start() {
    var authToken string

    authToken = o.authTokenFetcher.FetchAuthToken()
    log.Print("Starting SignalFx Firehose Nozzle...")
    o.setupFirehose(authToken)
    o.consumeFirehose()
    log.Print("SignalFx Firehose Nozzle shutting down...")
}

func (o *SignalFxFirehoseNozzle) Stop() {
    o.stop<- true
}

func (o *SignalFxFirehoseNozzle) setupFirehose(authToken string) {
    o.consumer = consumer.New(
        o.config.TrafficControllerURL,
        &tls.Config{InsecureSkipVerify: o.config.InsecureSSLSkipVerify},
        nil)
    o.consumer.SetIdleTimeout(time.Duration(o.config.FirehoseIdleTimeoutSeconds) * time.Second)
    o.messages, o.errs = o.consumer.Firehose(o.config.FirehoseSubscriptionID, authToken)
}

func (o *SignalFxFirehoseNozzle) consumeFirehose() {
    ticker := time.NewTicker(time.Duration(o.config.FlushIntervalSeconds) * time.Second)

    for {
        select {
        case <-o.stop:
            return
        case <-ticker.C:
            o.pushMetrics()
        case envelope := <-o.messages:
            dps := o.datapointsFromEnvelope(envelope)
            for i, _ := range dps {
                if o.shouldShipDatapoint(dps[i]) {
                    o.datapointBuffer = append(o.datapointBuffer, dps[i])
                }
            }
        case err := <-o.errs:
            o.handleError(err)
            o.pushMetrics()
        }
    }
}

func (o *SignalFxFirehoseNozzle) pushMetrics() {
    if len(o.datapointBuffer) == 0 {
        return
    }

    log.Printf("Pushing %d Firehose metrics to SignalFx", len(o.datapointBuffer))

    err := o.client.AddDatapoints(context.Background(), o.datapointBuffer)
    if err != nil {
        log.Print("Error shipping firehose datapoints to SignalFx: ", err)
        // If there is an error sending datapoints then just forget about them.
    }
    o.datapointBuffer = o.datapointBuffer[:0]
}

func (o *SignalFxFirehoseNozzle) handleError(err error) {
    log.Printf("Closing connection with traffic controller due to %v", err)
    o.consumer.Close()

    time.Sleep(time.Duration(o.config.FirehoseReconnectDelaySeconds) * time.Second)

    log.Println("Reconnecting to Firehose")

    o.setupFirehose(o.authTokenFetcher.FetchAuthToken())
}

// The ContainerMetric envelopes contain multiple metrics per envelope.  The
// rest are 1:1.
func (o *SignalFxFirehoseNozzle) datapointsFromEnvelope(envelope *events.Envelope) []*datapoint.Datapoint {
    eventType := envelope.GetEventType()
    origin := envelope.GetOrigin()

    dimensions := map[string]string {
        "job": envelope.GetJob(),
        "deployment": envelope.GetDeployment(),
        "host": envelope.GetIp(),
        // "index" in the firehose is a long guid value that indicates the BOSH
		// instance id, whereas in BOSH HM metrics, it is a simple cardinal #
		// indicating the instance index (e.g. 0, 1, 2, etc.).  Call the
		// firehose "index" the same as the BOSH HM "id" field for consistency.
		// They appear to be the same thing.
		// Also, don't just call it "id" since that is a reserved dimension
		// name in the backend.
        "bosh_id": envelope.GetIndex(),
        "metric_source": "cloudfoundry",
    }

	properties := map[string]string{}

    for k, v := range envelope.GetTags() {
        dimensions[k] = v
    }

    ts := time.Unix(0, envelope.GetTimestamp())

    switch eventType {
    case events.Envelope_ContainerMetric:
        contMetric := envelope.GetContainerMetric()
        guid := contMetric.GetApplicationId()

        dimensions["app_instance_index"] = strconv.Itoa(int(contMetric.GetInstanceIndex()))
        dimensions["app_id"] = guid

        // Send app metadata as both dims and properties since navigator views
        // seem to really want them as properties.
        dimensions["app_name"] = o.metadataFetcher.GetAppNameForGUID(guid)
        dimensions["app_space"] = o.metadataFetcher.GetSpaceNameForGUID(guid)
        dimensions["app_org"] = o.metadataFetcher.GetOrgNameForGUID(guid)

        return makeContainerDatapoints(dimensions, properties, ts, contMetric)
    case events.Envelope_ValueMetric:
        valueMetric := envelope.GetValueMetric()
        return []*datapoint.Datapoint {
            NewDatapointWithProps(origin + "." + valueMetric.GetName(),
                      dimensions,
					  properties,
                      datapoint.NewFloatValue(valueMetric.GetValue()),
                      datapointType(origin, valueMetric.GetName(), datapoint.Gauge),
                      ts),
        }
    case events.Envelope_CounterEvent:
        counterMetric := envelope.GetCounterEvent()
        return []*datapoint.Datapoint {
            NewDatapointWithProps(origin + "." + counterMetric.GetName(),
                      dimensions,
					  properties,
                      datapoint.NewIntValue(int64(counterMetric.GetTotal())),
                      datapointType(origin, counterMetric.GetName(), datapoint.Counter),
                      ts),
        }
    // TODO: see if there are any metrics we could pull out of these
    case events.Envelope_HttpStartStop:
        return []*datapoint.Datapoint{}
    // TODO: figure out what these could be and derive metrics if applicable
    case events.Envelope_Error:
        return []*datapoint.Datapoint{}
    case events.Envelope_LogMessage:
        return []*datapoint.Datapoint{}
    default:
        log.Printf("Unknown envelope type: %s", eventType)
        return []*datapoint.Datapoint{}
    }
}

func makeContainerDatapoints(dimensions map[string]string,
                             properties map[string]string,
                             timestamp time.Time,
                             contMetric *events.ContainerMetric) []*datapoint.Datapoint {
    return []*datapoint.Datapoint {
        NewDatapointWithProps("container.cpu_percentage",
            dimensions,
			properties,
            datapoint.NewFloatValue(contMetric.GetCpuPercentage()),
            datapoint.Gauge,
            timestamp),
        NewDatapointWithProps("container.memory_bytes",
            dimensions,
			properties,
            datapoint.NewIntValue(int64(contMetric.GetMemoryBytes())),
            datapoint.Gauge,
            timestamp),
        NewDatapointWithProps("container.memory_bytes_quota",
            dimensions,
			properties,
            datapoint.NewIntValue(int64(contMetric.GetMemoryBytesQuota())),
            datapoint.Gauge,
            timestamp),
        NewDatapointWithProps("container.disk_bytes",
            dimensions,
			properties,
            datapoint.NewIntValue(int64(contMetric.GetDiskBytes())),
            datapoint.Gauge,
            timestamp),
        NewDatapointWithProps("container.disk_bytes_quota",
            dimensions,
			properties,
            datapoint.NewIntValue(int64(contMetric.GetDiskBytesQuota())),
            datapoint.Gauge,
            timestamp),
    }
}
