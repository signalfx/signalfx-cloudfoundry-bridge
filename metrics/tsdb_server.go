package metrics

// This code is adapted from the bosh-hm-forwarder app at
// https://github.com/cloudfoundry/bosh-hm-forwarder
// Basically, we need to act as a TSDB "Telnet" server that Bosh HM sends VM
// metrics to

import (
    "fmt"
    "log"
    "time"
    "strings"
    "strconv"

    "github.com/cloudfoundry/bosh-hm-forwarder/tcp"
    "github.com/signalfx/golib/datapoint"
    "golang.org/x/net/context"
)

// This is the port that the BOSH HM OpenTSDB plugin is configured to connect
// to in PCF.  The JMX Bridge component has this hard coded so it should be
// safe to do the same here.
const tsdbPort = 13321

const initialBufferCapacity = 1000

// We accept port purely so we don't muck with global variables in testing
func StartTSDBServer(client SignalFxClient, flushInterval int, port int) (error) {
    if port == 0 {
        port = tsdbPort
    }

    err := tcp.Open(port, startMessageHandler(client, flushInterval))
    if err != nil {
        log.Print("Could not open TSDB server port", err)
    }
    return err
}

func startMessageHandler(client SignalFxClient, flushInterval int) chan<- string {
    // Buffer the channel so that the channel hopefully won't ever block when
    // the metrics are in the process of being shipped to the ingest API.
    dataCh := make(chan string, initialBufferCapacity)

    go handleMessages(dataCh, client, flushInterval)
    return dataCh
}


func handleMessages(tsdbLines chan string, client SignalFxClient, flushInterval int) {
    ticker := time.NewTicker(time.Second * time.Duration(flushInterval))
    defer ticker.Stop()

    datapointBuffer := make([]*datapoint.Datapoint, 0, initialBufferCapacity)

    var message string
    for {
        select {
        case message = <-tsdbLines:
            dp, err := buildDatapoint(message)
            if err != nil {
                continue
            }

            datapointBuffer = append(datapointBuffer, dp)
        case <-ticker.C:
            // Just send the datapoints synchronously for now since the data channel can buffer
            err := client.AddDatapoints(context.Background(), datapointBuffer)

            // Right now if there is an error shipping the datapoints to
            // SignalFx, we just forget about them and move on.  Some other
            // possiblities are: letting the buffer grow and just retry sending
            // them on the next tick; or immediately retrying one (or a fixed
            // number of) additional times.
            if err != nil {
                log.Println("Error shipping datapoints: ", err)
            }

            // Old datapoints will be GC'd as they are overwritten in the
            // backing array of the slice.  Conceivably, if one interval had an
            // abnormally large number of metrics that caused the buffer to
            // expand a lot, those datapoints might not be GC'd ever if the
            // buffer never filled that much again to overwrite them in the
            // backing array.  This should be fine since the total memory usage
            // (post-GC) of the buffer would never exceed that of the busiest
            // interval.  If this proves a problem, the simplest solution would
            // be to just recreate the datapointBuffer slice from scratch and
            // let the old elements be GC'd.
            datapointBuffer = datapointBuffer[:0]
        }
    }
}

func buildMap(tokens []string, startAt int) map[string]string {
    parsed := make(map[string]string)

    for i := startAt; i < len(tokens); i++ {
        token := tokens[i]
        split := strings.Split(token, "=")
        value := ""
        if len(split) > 1 {
            value = split[1]
        }
        parsed[split[0]] = value
    }
    return parsed
}

func buildDatapoint(message string) (*datapoint.Datapoint, error) {
    tokens := strings.Split(message, " ")

    if len(tokens) < 4 {
        return nil, fmt.Errorf("Malformed TSDB message: %s", message)
    }

    metricName := tokens[1]
    secondsSinceEpoch, err := strconv.ParseInt(tokens[2], 10, 64)
    if err != nil {
        log.Println("Cannot parse message: ", err)
        return nil, err
    }
    value, err := strconv.ParseFloat(tokens[3], 64)
    if err != nil {
        log.Println("Cannot parse message: ", err)
        return nil, err
    }
    dimensions := buildMap(tokens, 4)

    dimensions["metric_source"] = "cloudfoundry"

    return datapoint.New(metricName,
                         dimensions,
                         datapoint.NewFloatValue(value),
                         datapoint.Gauge,
                         time.Unix(secondsSinceEpoch, 0)), nil
}
