package com.signalfx.cloudfoundry;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.util.Map;

import static com.google.common.base.Preconditions.checkNotNull;

/**
 * Store CloudFoundry platform configuration values.
 */
public class Configuration {
    private static final String JMX_IP = "JMX_IP";
    private static final String JMX_USERNAME = "JMX_USERNAME";
    private static final String JMX_PASSWORD = "JMX_PASSWORD";
    private static final String JMX_SSL_ENABLED = "JMX_SSL_ENABLED";
    private static final String SFX_ACCESS_KEY = "SFX_ACCESS_KEY";
    private static final String SFX_INGEST_URL = "SFX_INGEST_URL";
    private static final String JMX_SSL_CERT_SOURCE = "JMX_SSL_CERT_SOURCE";
    private static final String JMX_SSL_CERT = "JMX_SSL_CERT";

    @JsonProperty(value = JMX_IP)
    private String jmxIp;

    @JsonProperty(value = JMX_USERNAME)
    private String jmxUsername;

    @JsonProperty(value = JMX_PASSWORD)
    private String jmxPassword;

    @JsonProperty(value = SFX_ACCESS_KEY)
    private String sfxAccessKey;

    @JsonProperty(value = SFX_INGEST_URL)
    private String sfxIngestUrl;

    @JsonProperty(value = JMX_SSL_ENABLED)
    private boolean jmxSslEnabled;

    @JsonProperty(value = JMX_SSL_CERT_SOURCE)
    private CertificateSource jmxSslCertSource;

    @JsonProperty(value = JMX_SSL_CERT)
    private String jmxSslCert;

    public static Configuration fromMap(Map<String, String> env) throws IOException {
        ObjectMapper objectMapper = new ObjectMapper().configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
        String jsonEnv = objectMapper.writeValueAsString(env);
        Configuration conf = objectMapper.readValue(jsonEnv, Configuration.class);

        // Could use something like javax.validation.* but probably overkill for now.
        checkValid(JMX_IP, conf.jmxIp);
        checkValid(JMX_USERNAME, conf.jmxUsername);
        checkValid(JMX_PASSWORD, conf.jmxPassword);
        checkValid(SFX_ACCESS_KEY, conf.sfxAccessKey);
        checkValid(SFX_INGEST_URL, conf.sfxIngestUrl);

        if (conf.isJmxSslEnabled()) {
            checkNotNull(conf.jmxSslCertSource, "JMX_SSL_CERT_SOURCE is null");
            checkValid(Configuration.JMX_SSL_CERT, conf.jmxSslCert);
        }

        return conf;
    }

    private static void checkValid(String key, String value) {
        if (value == null || value.length() == 0) {
            throw new ConfigurationException(
                    String.format("Invalid configuration option %s with value %s", key, value));
        }
    }

    public String getSfxIngestUrl() {
        return sfxIngestUrl;
    }

    public String getJmxIp() {
        return jmxIp;
    }

    public String getJmxUsername() {
        return jmxUsername;
    }

    public String getJmxPassword() {
        return jmxPassword;
    }

    public CertificateSource getJmxSslCertSource() {
        return jmxSslCertSource;
    }

    public String getJmxSslCert() {
        return jmxSslCert;
    }

    public String getSfxAccessKey() {
        return sfxAccessKey;
    }

    public boolean isJmxSslEnabled() {
        return jmxSslEnabled;
    }
}
