package com.signalfx.cloudfoundry;

/**
 * Exception for rethrowing checked exceptions from streams.
 */
public class RethrowException extends RuntimeException {
    private final Exception exception;

    public RethrowException(Exception exception) {
        this.exception = exception;
    }

    public Exception getException() {
        return exception;
    }
}
