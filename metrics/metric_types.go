package metrics

import (
    //"log"
    "github.com/signalfx/golib/datapoint"
)

var metricNameToType = map[string]map[string]datapoint.MetricType{
    "cc": map[string]datapoint.MetricType{
        "http_status.1XX": datapoint.Counter,
        "http_status.2XX": datapoint.Counter,
        "http_status.3XX": datapoint.Counter,
        "http_status.4XX": datapoint.Counter,
        "http_status.5XX": datapoint.Counter,
    },
    "uaa": map[string]datapoint.MetricType{
        "audit_service.client_authentication_count": datapoint.Counter,
        "audit_service.client_authentication_failure_count": datapoint.Counter,
        "audit_service.principal_authentication_failure_count": datapoint.Counter,
        "audit_service.principal_not_found_count": datapoint.Counter,
        "audit_service.user_authentication_count": datapoint.Counter,
        "audit_service.user_authentication_failure_count": datapoint.Counter,
        "audit_service.user_not_found_count": datapoint.Counter,
        "audit_service.user_password_failures": datapoint.Counter,
    },
}

//var metricTypeToHuman = map[datapoint.MetricType]string{
    //datapoint.Counter: "cumulative_counter",
    //datapoint.Gauge: "gauge",
//}

func datapointType(origin string, metricName string, defaultType datapoint.MetricType) datapoint.MetricType {
    originMap := metricNameToType[origin]
    if originMap != nil {
        metricType, ok := originMap[metricName]
        if ok {
            return metricType
        }
    }
    return defaultType
}
