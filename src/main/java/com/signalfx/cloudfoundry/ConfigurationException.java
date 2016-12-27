package com.signalfx.cloudfoundry;


/**
 * Thrown when there is a problem loading/parsing configuration.
 */
public class ConfigurationException extends RuntimeException {
    public ConfigurationException(String message) {
        super(message);
    }
}
