package com.signalfx.cloudfoundry;

import com.google.common.collect.ImmutableMap;
import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.management.*;
import java.io.IOException;
import java.util.*;
import java.util.stream.Collectors;

import static com.signalfx.cloudfoundry.Utils.formatMultilineList;
import static com.signalfx.cloudfoundry.Utils.formatMultilineMap;
import static com.signalfx.cloudfoundry.Utils.nullishStringToNull;

/**
 * Responsible for discovering metrics via JMX that are ready to later be probed.
 */
public class Discovery {
    private final static Logger logger = LoggerFactory.getLogger(Discovery.class);

    private static class PlatformDiscoverResult {
        private final List<Discovered> discovered;
        private final Map<String, String> ipCache;

        public PlatformDiscoverResult(List<Discovered> discovered, Map<String, String> ipCache) {
            this.discovered = discovered;
            this.ipCache = ipCache;
        }

        public List<Discovered> getDiscovered() {
            return discovered;
        }

        public Map<String, String> getIpCache() {
            return ipCache;
        }
    }

    private static List<ObjectName> getSystemObjects(MBeanServerConnection mbsc) throws MalformedObjectNameException, IOException {
        return mbsc.queryNames(new ObjectName("org.cloudfoundry:*"), null).stream().filter(objectName -> {
            String deployment = objectName.getKeyProperty("deployment");
            return deployment != null && !deployment.equals("cf");
        }).collect(Collectors.toList());
    }

    public List<Discovered> discover(MBeanServerConnection mbsc) throws Exception {
        logger.info("Discovering metrics...");

        final List<Discovered> discovered = new ArrayList<>();

        PlatformDiscoverResult platformDiscoverResult = discoverPlatformMetrics(mbsc);
        final Map<String, String> ipCache = platformDiscoverResult.getIpCache();
        logger.info("Loaded IP cache:\n{}", formatMultilineMap(ipCache));

        discovered.addAll(platformDiscoverResult.getDiscovered());
        discovered.addAll(discoverSystemMetrics(mbsc, ipCache));

        logger.info("Discovered {} metric sources:\n{}", discovered.size(), Utils.formatMultilineList(discovered));

        return discovered;
    }

    private List<Discovered> discoverSystemMetrics(MBeanServerConnection mbsc, Map<String, String> ipCache) throws Exception {
        try {
            return getSystemObjects(mbsc).stream()
                    .map(objectName -> {
                        String ip = nullishStringToNull(objectName.getKeyProperty("ip"));
                        String job = nullishStringToNull(objectName.getKeyProperty("job"));

                        if (ip == null) {
                            MBeanInfo mbeanInfo = null;
                            try {
                                mbeanInfo = mbsc.getMBeanInfo(objectName);
                            } catch (JMException e) {
                                throw new RethrowException(e);
                            } catch (IOException e) {
                                throw new RethrowException(e);
                            }
                            final String id = (String) mbeanInfo.getDescriptor().getFieldValue("id");
                            if (id != null) {
                                ip = ipCache.get(id);
                            }
                        }

                        if (ip != null && job != null) {
                            return new Discovered(objectName, ip, job, JobMetrics.systemAttributeToMetricType);
                        } else {
                            logger.warn("ip and job property missing for {}.", objectName);
                            return null;
                        }
                    })
                    .filter(Objects::nonNull)
                    .collect(Collectors.toList());
        } catch (RethrowException e) {
            throw e.getException();
        }
    }

    private PlatformDiscoverResult discoverPlatformMetrics(MBeanServerConnection mbsc) throws MalformedObjectNameException, IOException {
        final Map<String, String> ipCache = new HashMap<>();

        return new PlatformDiscoverResult(
                mbsc.queryNames(new ObjectName("org.cloudfoundry:deployment=cf,job=*,index=*,ip=*"), null).stream()
                        .map(objectName -> {
                            final String ip = nullishStringToNull(objectName.getKeyProperty("ip"));
                            final String index = nullishStringToNull(objectName.getKeyProperty("index"));
                            final String job = nullishStringToNull(objectName.getKeyProperty("job"));

                            if (index != null && ip != null) {
                                ipCache.put(index, ip);
                            }

                            if (ip != null && job != null) {
                                ImmutableMap<String, SignalFxProtocolBuffers.MetricType> attributes = JobMetrics.mapper.get(job);
                                if (attributes != null) {
                                    return new Discovered(objectName, ip, job, attributes);
                                } else {
                                    logger.info("No attribute mapping present for job type {}.", job);
                                }
                            } else {
                                logger.warn("ip and job property missing for {}.", objectName);
                            }

                            return null;
                        })
                        .filter(Objects::nonNull)
                        .collect(Collectors.toList()), ipCache);
    }
}
