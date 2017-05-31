package metrics

import (
    "golang.org/x/net/context"
    "log"
	"os"
    "time"

    "github.com/signalfx/golib/datapoint"
)

type SignalFxClient interface {
    AddDatapoints(context.Context, []*datapoint.Datapoint) error
}


var DEBUG = os.Getenv("DEBUG") != ""

func DebugLog(format string, args ...interface{}) {
	if DEBUG {
		log.Printf(format, args...)
	}
}

func NewDatapointWithProps(metric string,
                           dimensions map[string]string,
                           properties map[string]string,
                           value datapoint.Value,
                           metricType datapoint.MetricType,
                           timestamp time.Time) *datapoint.Datapoint {
    dp := datapoint.New(metric, dimensions, value, metricType, timestamp)
    for k, v := range properties {
        dp.SetProperty(k, v)
    }
    return dp
}
