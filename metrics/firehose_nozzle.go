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
    config                *Config
    errs                  <-chan error
    messages              <-chan *events.Envelope
    authTokenFetcher      AuthTokenFetcher
    consumer              *consumer.Consumer
    client                SignalFxClient
    stop                   chan bool
    datapointBuffer       []*datapoint.Datapoint
    totalMessagesReceived int
    metadataFetcher       *AppMetadataFetcher
}

type AuthTokenFetcher interface {
    FetchAuthToken() string
}

func NewSignalFxFirehoseNozzle(config *Config,
                               tokenFetcher AuthTokenFetcher,
                               client SignalFxClient,
                               metadataFetcher *AppMetadataFetcher) *SignalFxFirehoseNozzle {
    return &SignalFxFirehoseNozzle{
        config:           config,
        client:           client,
        errs:             make(<-chan error),
        messages:         make(<-chan *events.Envelope),
        stop:             make(chan bool),
        authTokenFetcher: tokenFetcher,
        datapointBuffer:  make([]*datapoint.Datapoint, 0, 1000),
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
            o.datapointBuffer = append(o.datapointBuffer, dps...)
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
    log.Printf("Pushing %d metrics to SignalFx", len(o.datapointBuffer))

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
        "index": envelope.GetIndex(),
        "metric_source": "cloudfoundry",
    }

    ts := time.Unix(0, envelope.GetTimestamp())

    switch eventType {
    case events.Envelope_ContainerMetric:
        contMetric := envelope.GetContainerMetric()
        guid := contMetric.GetApplicationId()
        dimensions["app_id"] = guid
        dimensions["app_instance_index"] = strconv.Itoa(int(contMetric.GetInstanceIndex()))
        dimensions["app_name"] = o.metadataFetcher.GetAppNameForGUID(guid)
        dimensions["app_space"] = o.metadataFetcher.GetSpaceNameForGUID(guid)
        dimensions["app_org"] = o.metadataFetcher.GetOrgNameForGUID(guid)

        log.Print(dimensions)

        return makeContainerDatapoints(dimensions, ts, contMetric)
    case events.Envelope_ValueMetric:
        valueMetric := envelope.GetValueMetric()
        return []*datapoint.Datapoint {
            datapoint.New(origin + "." + valueMetric.GetName(),
                      dimensions,
                      datapoint.NewFloatValue(valueMetric.GetValue()),
                      datapoint.Gauge,
                      ts),
        }
    case events.Envelope_CounterEvent:
        counterMetric := envelope.GetCounterEvent()
        return []*datapoint.Datapoint {
            datapoint.New(origin + "." + counterMetric.GetName(),
                      dimensions,
                      datapoint.NewIntValue(int64(counterMetric.GetTotal())),
                      datapoint.Counter,
                      ts),
        }
    // TODO: see if there are any metrics we could pull out of these
    case events.Envelope_HttpStartStop:
        return []*datapoint.Datapoint {}
    // TODO: figure out what these could be and derive metrics if applicable
    case events.Envelope_Error:
        return []*datapoint.Datapoint {}
    default:
        log.Printf("Unknown envelope type: %#v", envelope)
        return []*datapoint.Datapoint {}
    }
}

func makeContainerDatapoints(dimensions map[string]string,
                             timestamp time.Time,
                             contMetric *events.ContainerMetric) []*datapoint.Datapoint {
    return []*datapoint.Datapoint {
        datapoint.New("cpu_percentage",
            dimensions,
            datapoint.NewFloatValue(contMetric.GetCpuPercentage()),
            datapoint.Gauge,
            timestamp),
        datapoint.New("memory_bytes",
            dimensions,
            datapoint.NewIntValue(int64(contMetric.GetMemoryBytes())),
            datapoint.Gauge,
            timestamp),
        datapoint.New("memory_bytes_quota",
            dimensions,
            datapoint.NewIntValue(int64(contMetric.GetMemoryBytesQuota())),
            datapoint.Gauge,
            timestamp),
        datapoint.New("disk_bytes",
            dimensions,
            datapoint.NewIntValue(int64(contMetric.GetDiskBytes())),
            datapoint.Gauge,
            timestamp),
        datapoint.New("disk_bytes_quota",
            dimensions,
            datapoint.NewIntValue(int64(contMetric.GetDiskBytesQuota())),
            datapoint.Gauge,
            timestamp),
    }
}
