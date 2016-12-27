package com.signalfx.cloudfoundry;

import com.signalfx.endpoint.SignalFxEndpoint;
import com.signalfx.endpoint.SignalFxReceiverEndpoint;
import com.signalfx.metrics.auth.StaticAuthToken;
import com.signalfx.metrics.connection.HttpDataPointProtobufReceiverFactory;
import com.signalfx.metrics.connection.HttpEventProtobufReceiverFactory;
import com.signalfx.metrics.errorhandler.MetricError;
import com.signalfx.metrics.errorhandler.OnSendErrorHandler;
import com.signalfx.metrics.flush.AggregateMetricSender;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.management.remote.JMXServiceURL;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.Collections;


public class Main {
    private final static Logger logger = LoggerFactory.getLogger(Main.class);

    public static void main(String[] args) {
        // Load configuration from environment.
        final Configuration conf;
        try {
            conf = Configuration.fromMap(System.getenv());
        } catch (Exception e) {
            logger.error("Failed to load configuration.", e);
            System.exit(1);
            return;
        }

        // Possibly load configured SSL certificate.
        try {
            if (conf.isJmxSslEnabled()) {
                SSL.configureSsl(conf);
            }
        } catch (Exception e) {
            logger.error("Failed to initialize SSL certificates.", e);
            System.exit(1);
            return;
        }

        // SignalFx configuration.
        final URL url;
        try {
            url = new URL(conf.getSfxIngestUrl());
        } catch (MalformedURLException e) {
            logger.error("Failed to parse ingestion URL {}: {}", conf.getSfxIngestUrl(), e.getMessage());
            System.exit(1);
            return;
        }

        final SignalFxReceiverEndpoint signalFxEndpoint = new SignalFxEndpoint(url.getProtocol(), url.getHost(), url.getPort());
        final AggregateMetricSender metricSender = new AggregateMetricSender("CloudFoundry",
                new HttpDataPointProtobufReceiverFactory(signalFxEndpoint).setVersion(2),
                new HttpEventProtobufReceiverFactory(signalFxEndpoint),
                new StaticAuthToken(conf.getSfxAccessKey()),
                Collections.<OnSendErrorHandler>singleton(new OnSendErrorHandler() {
                    @Override
                    public void handleError(MetricError metricError) {
                        logger.error("Unable to POST metrics: {}", metricError.getMessage());
                    }
                }));

        // JMX configuration.
        String serviceUrlString = String.format("service:jmx:rmi:///jndi/rmi://%s:44444/jmxrmi", conf.getJmxIp());
        final JMXServiceURL serviceUrl;
        try {
            serviceUrl = new JMXServiceURL(serviceUrlString);
        } catch (MalformedURLException e) {
            logger.error("Invalid service URL {}: {}", serviceUrlString, e.getMessage());
            System.exit(1);
            return;
        }

        // Kick off the main loop.
        Runner runner = new Runner(conf, serviceUrl, metricSender);
        Runtime.getRuntime().addShutdownHook(new Thread(runner::stop));
        runner.run();
    }
}
