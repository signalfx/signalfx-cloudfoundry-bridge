package metrics

import (
    "strings"

    "github.com/signalfx/golib/v3/datapoint"
)

// Filters datapoints based on deployment name and metric name
type MetricFilter struct {
    // An optimization to quickly look up whether to allow a deployment
    deploymentSet     map[string]bool
    // Similar to the above but a blacklist on metric name
    metricBlacklistSet map[string]bool
}

func NewMetricFilter(config *Config) *MetricFilter {
    deploymentSet := make(map[string]bool)
    for _, v := range config.DeploymentsToInclude {
        deploymentSet[strings.TrimSpace(v)] = true
    }

    metricsBlacklistSet := make(map[string]bool)
    for _, v := range config.MetricsToExclude {
        metricsBlacklistSet[strings.TrimSpace(v)] = true
    }

    return &MetricFilter{
        deploymentSet: deploymentSet,
        metricBlacklistSet: metricsBlacklistSet,
    }
}

// It is far simpler to filter the already created datapoint.  We could
// short-cut the filtering and do it on the firehose envelope or on the TSDB
// line before datapoints are created and save a bit of time, but this should
// be sufficient for now.
func (o *MetricFilter) shouldShipDatapoint(dp *datapoint.Datapoint) bool {
    deploymentAllowed := len(o.deploymentSet) == 0 || o.deploymentSet[dp.Dimensions["deployment"]]
    metricAllowed := !o.metricBlacklistSet[dp.Metric]

    return deploymentAllowed && metricAllowed
}
