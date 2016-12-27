package com.signalfx.cloudfoundry;

import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.management.Attribute;
import javax.management.InstanceNotFoundException;
import javax.management.MBeanServerConnection;
import javax.management.ReflectionException;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.regex.Pattern;

/**
 * Queries predefined metrics and submits them to SignalFx.
 */
public class Collection {
    private final static Logger logger = LoggerFactory.getLogger(Collection.class);
    private static Pattern metricNamePattern = Pattern.compile("^opentsdb\\.nozzle\\.");

    public static List<SignalFxProtocolBuffers.DataPoint> collect(Discovered discovered, MBeanServerConnection mbsc)
            throws InstanceNotFoundException, IOException, ReflectionException {
        logger.info("Collecting metrics from {}.", discovered);

        final List<SignalFxProtocolBuffers.DataPoint> datapoints = new ArrayList<>();
        final List<Attribute> attrs = mbsc.getAttributes(discovered.getObjectName(), discovered.getAttributes().keySet().toArray(new String[0])).asList();

        if (attrs.stream().allMatch(a -> a.getValue() == null)) {
            // This way we don't end up logging each one if they're all missing.
            logger.warn("No expected attributes are present for {}, skipping.", discovered.getObjectName());
            return datapoints;
        }

        for (Attribute attr : attrs) {
            if (attr.getValue() == null) {
                logger.warn("{} has no attribute value for {}", discovered.getObjectName(), attr.getName());
                continue;
            }
            SignalFxProtocolBuffers.MetricType metricType = discovered.getAttributes().get(attr.getName());

            if (metricType == null) {
                throw new RuntimeException("No metric type for " + attr.getName());
            }

            // Strip the metric name if it starts with "opentsdb.nozzle."
            final String metricName = metricNamePattern.matcher(attr.getName()).replaceFirst("");

            SignalFxProtocolBuffers.DataPoint.Builder dataPoint = SignalFxProtocolBuffers.DataPoint.newBuilder()
                    .setMetric(metricName)
                    .setMetricType(metricType)
                    .setValue(SignalFxProtocolBuffers.Datum.newBuilder()
                            .setDoubleValue((Double) attr.getValue()))
                    .addDimensions(
                            SignalFxProtocolBuffers.Dimension.newBuilder()
                                    .setKey("job")
                                    .setValue(discovered.getJob()))
                    .addDimensions(
                            SignalFxProtocolBuffers.Dimension.newBuilder()
                                    .setKey("metric_source")
                                    .setValue("cloudfoundry")
                    );
            if (discovered.getIp() != null) {
                dataPoint.addDimensions(
                        SignalFxProtocolBuffers.Dimension.newBuilder()
                                .setKey("host")
                                .setValue(discovered.getIp()));
            }

            datapoints.add(dataPoint.build());
        }

        return datapoints;
    }
}
