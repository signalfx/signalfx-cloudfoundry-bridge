# Overview
This integration is used to report Cloud Foundry platform metrics to SignalFx. The integration is composed of a Java agent that is built into a Pivotal tile and deployed via Ops Manager to Pivotal Cloud Foundry.

# Build
## Agent
To build the agent you can run `gradle` directly or to run a default build run `make jar`. This agent can be run locally if you can talk to the JMX Bridge.

Assuming the [required environmental](#Configuration) variables have been set run the agent with:

`java -jar cf-agent-$VERSION-all.jar`

## Tile
To build the tile install the tile tools with `pip install tile-generator`. See [the Pivotal documentation](https://docs.pivotal.io/tiledev/tile-generator.html) for details.

To build the tile run `make tile`.

If you want to be able to push the newly built tile directly to Cloud Foundry configure `pcf` [according to the documentation](https://docs.pivotal.io/tiledev/pcf-command.html).

You can now run `make build-and-push` to build and push the tile to your Cloud Foundry instance. You'll then need to apply your changes through Ops Manager to deploy the new version.

# Configuration
The agent is configured by environment variables. Required configuration values:

<dl>
    <dt>JMX_USERNAME</dt>
    <dd>JMX password username</dd>

    <dt>JMX_PASSWORD</dt>
    <dd>JMX Bridge password</dd>

    <dt>SFX_ACCESS_KEY</dt>
    <dd>SignalFx access token</dd>

    <dt>JMX_SSL_CERT_SOURCE</dt>
    <dd>Either FILE to read the JMX SSL cert from a file as specified in <tt>JMX_SSL_CERT</tt> or inline from the value of <tt>JMX_SSL_CERT</tt></dd>

    <dt>JMX_SSL_CERT</dt>
    <dd>See <tt>JMX_SSL_CERT_SOURCE</tt></dd>

    <dt>JMX_SSL_ENABLED</dt>
    <dd><tt>true</tt> if connecting to JMX over SSL, otherwise <tt>false</tt></dd>

    <dt>SFX_INGEST_URL</dt>
    <dd>Ingest URL for SignalFx</dd>

    <dt>JMX_IP</dt>
    <dd>IP address of the JMX Bridge</dd>
</dl>

These values are configured by the end user via the tile in Ops Manager.
