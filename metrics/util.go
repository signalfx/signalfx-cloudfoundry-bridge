package metrics

import (
	"log"
	"os"

	"golang.org/x/net/context"

	"github.com/signalfx/golib/v3/datapoint"
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
