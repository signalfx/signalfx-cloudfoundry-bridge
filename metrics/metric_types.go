package metrics

import (
	//"log"
    "github.com/signalfx/golib/datapoint"
)

var metricNameToType = map[string]map[string]datapoint.MetricType{
	"uaa": map[string]datapoint.MetricType{
		"audit_service.client_authentication_count": datapoint.Count,
		"audit_service.client_authentication_failure_count": datapoint.Count,
		"audit_service.principal_authentication_failure_count": datapoint.Count,
		"audit_service.principal_not_found_count": datapoint.Count,
		"audit_service.user_authentication_count": datapoint.Count,
		"audit_service.user_authentication_failure_count": datapoint.Count,
		"audit_service.user_not_found_count": datapoint.Count,
		"audit_service.user_password_failures": datapoint.Count,
	},
}

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
