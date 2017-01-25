package com.signalfx.cloudfoundry;

import com.google.common.collect.ImmutableMap;
import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;

import java.util.Map;

/**
 * Metrics that are gathered for each job.
 */
public class JobMetrics {
    private final static String DIEGO_CELL = "diego_cell";
    private final static String ROUTER = "router";
    private final static String CLOUD_CONTROLLER = "cloud_controller";
    private final static String UAA = "uaa";
    private final static String DOPPLER = "doppler";

    public static ImmutableMap<String, SignalFxProtocolBuffers.MetricType> systemAttributeToMetricType =
            ImmutableMap.<String, SignalFxProtocolBuffers.MetricType>builder()
                    .put("system.mem.percent", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.swap.percent", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.disk.ephemeral.percent", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.disk.system.percent", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.cpu.sys", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.cpu.user", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.cpu.wait", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("system.healthy", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .build();

    private static ImmutableMap<String, SignalFxProtocolBuffers.MetricType> diegoCellAttributeToMetricType =
            ImmutableMap.<String, SignalFxProtocolBuffers.MetricType>builder()
                    .put("opentsdb.nozzle.rep.CapacityTotalMemory", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.rep.CapacityRemainingMemory", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.rep.CapacityTotalDisk", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.rep.CapacityRemainingDisk", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.rep.ContainerCount", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.rep.UnhealthyCell", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .build();

    private static ImmutableMap<String, SignalFxProtocolBuffers.MetricType> routerAttributeToMetricType =
            ImmutableMap.<String, SignalFxProtocolBuffers.MetricType>builder()
                    .put("opentsdb.nozzle.gorouter.total_routes", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.gorouter.total_requests", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.gorouter.responses", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.gorouter.bad_gateways", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .build();

    private static ImmutableMap<String, SignalFxProtocolBuffers.MetricType> cloudControllerAttributeToMetricType =
            ImmutableMap.<String, SignalFxProtocolBuffers.MetricType>builder()
                    .put("opentsdb.nozzle.cc.requests.completed", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.cc.requests.outstanding", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.cc.tasks_running.count", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.cc.log_count.error", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.cc.log_count.fatal", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.cc.log_count.warn", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .build();

    private static ImmutableMap<String, SignalFxProtocolBuffers.MetricType> uaaAttributeToMetricType =
            ImmutableMap.<String, SignalFxProtocolBuffers.MetricType>builder()
                    .put("opentsdb.nozzle.uaa.audit_service.client_authentication_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.client_authentication_failure_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.principal_authentication_failure_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.principal_not_found_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.user_authentication_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.user_authentication_failure_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.user_not_found_count", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.user_password_changes", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .put("opentsdb.nozzle.uaa.audit_service.user_password_failures", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .build();

    private static ImmutableMap<String, SignalFxProtocolBuffers.MetricType> dopplerAttributeToMetricType =
            ImmutableMap.<String, SignalFxProtocolBuffers.MetricType>builder()
                    .put("opentsdb.nozzle.DopplerServer.memoryStats.numBytesAllocatedStack", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.DopplerServer.memoryStats.numBytesAllocatedHeap", SignalFxProtocolBuffers.MetricType.GAUGE)
                    .put("opentsdb.nozzle.DopplerServer.sentMessagesFirehose", SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER)
                    .build();

    public static ImmutableMap<String, ImmutableMap<String, SignalFxProtocolBuffers.MetricType>> mapper =
            ImmutableMap.<String, ImmutableMap<String, SignalFxProtocolBuffers.MetricType>>builder()
                    .put(DIEGO_CELL, diegoCellAttributeToMetricType)
                    .put(ROUTER, routerAttributeToMetricType)
                    .put(CLOUD_CONTROLLER, cloudControllerAttributeToMetricType)
                    .put(UAA, uaaAttributeToMetricType)
                    .put(DOPPLER, dopplerAttributeToMetricType)
                    .build();
}
