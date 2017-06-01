# Summary
The signalfx-bridge is a CF component which forwards metrics from the Loggregator Firehose and Bosh Health Monitor to a [SignalFX](https://www.signalfx.com) deployment.

# Architecture

There are two sources of data that this application will pull from: the
Loggregator Firehose, and the Bosh Health Monitor (HM) OpenTSDB plugin.  The
Bosh HM provides lower level VM metrics and the Firehose provides CF component and
container metrics.

# Setup
You must configure the Bosh HM OpenTSDB plugin by setting the *Metrics IP
Address* of the BOSH Director to the server where this app is running.  This
can be set in the web admin interface under: PCF Ops Manager Director tile ->
Settings tab -> Director Config.

# Configuration
The agent is configured by environment variables. Required configuration values:

<dl>
    <dt>SIGNALFX_ACCESS_TOKEN</dt>
    <dd>SignalFx access token</dd>

    <dt>SIGNALFX_INGEST_URL</dt>
    <dd>Ingest URL for SignalFx</dd>
</dl>

These values are configured by the end user via the tile in Ops Manager.

## CloudFoundry UAA User for Firehose Nozzle

The SignalFx firehose nozzle requires a UAA user who is authorized to access
the Loggregator firehose and the CloudController API (to pull app metadata).
You can add a user by editing your Cloud Foundry manifest to include the details
about this user under the `properties.uaa.clients` section. For example to add
a user `signalfx-firehose-nozzle`:

```
properties:
  uaa:
    clients:
      signalfx-bridge:
        access-token-validity: 1209600
        authorized-grant-types: client_credentials,refresh_token
        override: true
        secret: <password>
        scope: openid,oauth.approvals,doppler.firehose,cloud_controller.admin_read_only
        authorities: oauth.login,doppler.firehose,cloud_controller.admin_read_only
```

Or with the uaac CLI tool:

```sh
uaac client add \
  --name signalfx-bridge \
  --scope openid,oauth.approvals,doppler.firehose,cloud_controller.admin_read_only \
  --authorized_grant_types client_credentials,refresh_token \
  --secret <password> \
  --access_token_validity 1209600 \
  --authorities oauth.login,doppler.firehose,cloud_controller.admin_read_only
```

## BOSH Director Access
The Bridge needs access to the BOSH directors to correctly link BOSH Health
Metrics to IP addresses of the VMs (unfortunately the BOSH HM Metrics don't
include IP address).  See [Creating a BOSH
Client](https://docs.pivotal.io/pivotalcf/1-10/customizing/opsmanager-create-bosh-client.html)
on how to do this with Pivotal Cloud Foundry.  If you are using a non-Pivotal
CF deployments, steps will differ but probably be mostly the same.

# Build

## Tile
To build the tile install the tile tools with `pip install tile-generator`. See [the Pivotal documentation](https://docs.pivotal.io/tiledev/tile-generator.html) for details.

To build the tile run `make tile`.

If you want to be able to push the newly built tile directly to Cloud Foundry configure `pcf` [according to the documentation](https://docs.pivotal.io/tiledev/pcf-command.html).

You can now run `make build-and-push` to build and push the tile to your Cloud Foundry instance. You'll then need to apply your changes through Ops Manager to deploy the new version.


# Running

The signalfx nozzle uses a configuration file to obtain the firehose URL, and other configuration parameters. The firehose and the signalfx servers both require authentication -- the firehose requires a valid username/password and signalfx requires a valid API key.

You can start the firehose nozzle by executing:
```
go run main.go -config config/signalfx-firehose-nozzle.json
```


# Tests

You need [ginkgo](http://onsi.github.io/ginkgo/) and go 1.5+ to run the tests. The tests can be executed by:
```
go build
ginkgo -r

```

## Local Development
In our Pivotal CF deployment, I setup a tinyproxy instance on the Ops Manager
server port 8888 that allows HTTPS to the necessary ports (default is only 443 and 563).
I then forward local port 8888 to that remote server to use it as an HTTP/HTTPS
proxy without opening up the AWS firewall.  I also forward the BOSH HM Metrics that
are configured to come to the Ops Manager machine on port 13322 to port 13321
on my local machine.

```sh
ssh pcf-ops -R 13322:localhost:13321 -L 8888:localhost:8888 -N
```

Envvars to use tinyproxy forwarded to localhost:
```sh
export http_proxy=http://127.0.0.1:8888 https_proxy=http://127.0.0.1:8888
```
