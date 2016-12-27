package com.signalfx.cloudfoundry;

import com.google.common.collect.ImmutableMap;
import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;

import javax.management.ObjectName;

import static com.google.common.base.Preconditions.checkArgument;
import static com.google.common.base.Preconditions.checkNotNull;

public class Discovered {
    private final ObjectName objectName;
    private final String ip;
    private final String job;
    private final ImmutableMap<String, SignalFxProtocolBuffers.MetricType> attributes;

    public Discovered(ObjectName objectName, String ip, String job, ImmutableMap<String, SignalFxProtocolBuffers.MetricType> attributes) {
        checkNotNull(objectName, "objectName cannot be null");
        checkNotNull(job, "job name cannot be null");
        checkArgument(!job.isEmpty(), "job name cannot be empty");
        checkNotNull(attributes, "attributes mapping cannot be null");

        this.objectName = objectName;
        this.ip = ip;
        this.job = job;
        this.attributes = attributes;
    }

    public ObjectName getObjectName() {
        return objectName;
    }

    public String getIp() {
        return ip;
    }

    public String getJob() {
        return job;
    }

    public ImmutableMap<String, SignalFxProtocolBuffers.MetricType> getAttributes() {
        return attributes;
    }

    @Override
    public String toString() {
        return "Discovered{" +
                "objectName=" + objectName +
                ", ip='" + ip + '\'' +
                ", job='" + job + '\'' +
                '}';
    }
}
