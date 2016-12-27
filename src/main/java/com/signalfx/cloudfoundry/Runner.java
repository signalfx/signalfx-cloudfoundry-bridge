package com.signalfx.cloudfoundry;

import com.signalfx.metrics.flush.AggregateMetricSender;
import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.management.JMException;
import javax.management.MBeanServerConnection;
import javax.management.remote.JMXConnector;
import javax.management.remote.JMXConnectorFactory;
import javax.management.remote.JMXServiceURL;
import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

/**
 * Executes discovery and collection for the duration of the program.
 */
public class Runner {
    private static Logger logger = LoggerFactory.getLogger(Runner.class);

    private static final int DELAY_BETWEEN_COLLECTION = 30;
    private static final int DELAY_BETWEEN_RECONNECT = 30;
    private static final long DELAY_BETWEEN_REDISCOVERY = TimeUnit.MINUTES.toSeconds(5);

    private final Configuration conf;
    private final JMXServiceURL serviceUrl;
    private final AggregateMetricSender metricSender;
    private final Discovery discovery = new Discovery();

    private JMXConnector jmxConnector;
    private MBeanServerConnection mbsc;
    private List<Discovered> discovered;
    private long lastDiscovery = 0;

    /**
     * Stop the runner by closing the underlying JMX connection.
     */
    public void stop() {
        logger.info("Stopping runner.");
        close();
    }

    private enum State {
        // Initial state.
        DISCONNECTED,

        // State where everything is good and metrics are being collected.
        COLLECTING,

        // Teardown the connection and start over.
        RESETTING,

        // Discovering MBeans before collection starts.
        DISCOVERING,
    }

    public Runner(Configuration conf, JMXServiceURL serviceUrl, AggregateMetricSender metricSender) {
        this.conf = conf;
        this.serviceUrl = serviceUrl;
        this.metricSender = metricSender;
    }

    private State connect() {
        Map<String, Object> env = new HashMap<>();
        env.put(JMXConnector.CREDENTIALS, new String[]{conf.getJmxUsername(), conf.getJmxPassword()});

        try {
            if (jmxConnector == null) {
                logger.info("Connecting to {}.", serviceUrl);
                jmxConnector = JMXConnectorFactory.connect(serviceUrl, env);
                mbsc = jmxConnector.getMBeanServerConnection();
            }
            return State.DISCOVERING;
        } catch (Exception e) {
            logger.error("Failed to connect to JMX bridge.", e);
            return State.DISCONNECTED;
        }
    }

    /**
     * Close JMX connection if connected
     *
     * Synchronized because can be called by both stop (in shutdown thread)
     * and through main loop.
     * @return new state
     */
    private synchronized State close() {
        if (jmxConnector != null) {
            try {
                logger.info("Closing existing JMX connection.");
                jmxConnector.close();
            } catch (IOException e1) {
                logger.warn("Error while closing JMX connection.", e1);
            }
            // Set to null regardless of close() success.
            logger.info("Resetting JMX connection.");
            jmxConnector = null;
        }
        return State.DISCONNECTED;
    }

    /**
     * Execute main loop.
     */
    public void run() {
        // Basic state machine to handle reconnects and errors.
        //
        // Happy path is:
        //     DISCONNECTED -> DISCOVERING -> COLLECTING
        //
        // On error while already connected:
        //     COLLECTING -> RESETTING -> DISCONNECTED -> DISCOVERING -> COLLECTING (on reconnection success)
        //
        // Rediscovery:
        //     COLLECTING -> DISCOVERING -> COLLECTING
        State state = State.DISCONNECTED;

        while (true) {
            logger.debug("Current state is {}.", state);

            switch (state) {
                // Code executed or methods called from here should only issue state transitions and
                // not raise exceptions unless they're intended to be fatal.

                case RESETTING:
                    state = close();
                    logger.info("Waiting {} seconds before reconnecting.", DELAY_BETWEEN_RECONNECT);
                    try {
                        TimeUnit.SECONDS.sleep(DELAY_BETWEEN_RECONNECT);
                    } catch (InterruptedException e) {
                        logger.info("Sleep interrupted after JMX connect failure.");
                    }
                    break;
                case DISCONNECTED:
                    state = connect();

                    if (state == State.DISCONNECTED) {
                        logger.info("Waiting {} seconds before reconnecting.", DELAY_BETWEEN_RECONNECT);
                        try {
                            TimeUnit.SECONDS.sleep(DELAY_BETWEEN_RECONNECT);
                        } catch (InterruptedException e) {
                            logger.info("Sleep interrupted after JMX connect failure.");
                        }
                    }
                    break;
                case COLLECTING:
                    if (needsRediscovery()) {
                        final long timeElapsed = System.nanoTime() - lastDiscovery;
                        logger.info(String.format("Initiating rediscover after %1$.3f seconds.", timeElapsed / 1.0e+9));
                        state = State.DISCOVERING;
                        break;
                    }

                    try {
                        performCollection(metricSender, mbsc, discovered);
                    } catch (Exception e) {
                        logger.error("Collection failed, resetting connection.", e);
                        state = State.RESETTING;
                    }

                    if (state == State.COLLECTING) {
                        logger.info("Sleeping for {} seconds.", DELAY_BETWEEN_COLLECTION);
                        try {
                            TimeUnit.SECONDS.sleep(DELAY_BETWEEN_COLLECTION);
                        } catch (InterruptedException e) {
                            logger.info("Sleep interrupted between collections.");
                        }
                    }
                    break;
                case DISCOVERING:
                    try {
                        // Query JMX for the MBeans we're interested in and normalize their properties.
                        discovered = discovery.discover(mbsc);
                        state = State.COLLECTING;
                        lastDiscovery = System.nanoTime();
                    } catch (Exception e) {
                        // If discovery fails reset the connection and try again.
                        logger.error("Discovery failed.", e);
                        state = State.RESETTING;
                    }
                    break;
            }

            logger.debug("New state is {}.", state);
        }
    }

    /**
     * Determines if a rediscovery is due.
     * @return true if rediscovery needed, otherwise false
     */
    private boolean needsRediscovery() {
        final long diff = System.nanoTime() - lastDiscovery;

        if (diff < 0) {
            logger.warn("Time delta was unexpectedly less than zero: {}.", diff);
            return true;
        } else {
            return TimeUnit.NANOSECONDS.toSeconds(diff) >= DELAY_BETWEEN_REDISCOVERY;
        }
    }

    /**
     * Perform collection of discovered metrics
     * @param metricSender
     * @param mbsc
     * @param discovered
     * @throws Exception
     */
    private void performCollection(AggregateMetricSender metricSender, MBeanServerConnection mbsc, List<Discovered> discovered) throws Exception {
        final long startTime = System.nanoTime();
        final List<SignalFxProtocolBuffers.DataPoint> dataPoints;

        try {
            // Collect and send metrics.
            dataPoints = discovered.parallelStream().flatMap(disc -> {
                // Nastiness because streams don't handle checked exceptions.
                try {
                    return Collection.collect(disc, mbsc).stream();
                } catch (JMException e) {
                    throw new RethrowException(e);
                } catch (IOException e) {
                    throw new RethrowException(e);
                }
            }).collect(Collectors.toList());
        } catch (RethrowException e) {
            throw e.getException();
        }

        try (AggregateMetricSender.Session session = metricSender.createSession()) {
            for (SignalFxProtocolBuffers.DataPoint dataPoint : dataPoints) {
                session.setDatapoint(dataPoint);
            }
        }

        // Could be negative. Do anything about it?
        final long timeElapsed = System.nanoTime() - startTime;
        logger.info(String.format("Sync took %1$.3f seconds.", timeElapsed / 1.0e+9));
    }
}
