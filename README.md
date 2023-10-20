> # :warning: End of Support (EoS) Notice
> **The SignalFx Cloud Foundry Bridge and SignalFx Smart Agent have reached End of Support.**
>
> The [Splunk Distribution of OpenTelemetry Collector for Pivotal Cloud Foundry](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/cloudfoundry) is the successor.


>ℹ️&nbsp;&nbsp;SignalFx was acquired by Splunk in October 2019. See [Splunk SignalFx](https://www.splunk.com/en_us/investor-relations/acquisitions/signalfx.html) for more information.

# Summary
The signalfx-bridge is a CF component which forwards metrics from the
Loggregator Firehose and Bosh Health Monitor to a
[SignalFx](https://www.signalfx.com) deployment.

**Note**: As of PCF 2.0, Bosh VM Health Metrics are emitted by the loggregator and you will not need to enable the TSDB server, which previously served teh BOSH VM health metrics.

# Architecture

There are two sources of data that this application will pull from: the
Loggregator Firehose, and the Bosh Health Monitor (HM) OpenTSDB plugin.  The
Bosh HM provides lower level VM metrics and the Firehose provides CF component and
container metrics.

## Setup (Pivotal CF)
Install the Pivotal Tile from the Pivotal Network.

Then, you must set the *Metrics IP Address* setting to the server where this app is
running.  This can be set in the web admin interface under: PCF Ops Manager
Director tile -> Settings tab -> Director Config.

## Setup (Other CF)
Install this app through a [BOSH release]().

You must install the BOSH HM TSDB plugin and include the following
configuration in order to get VM health metrics:

```yaml
...
"plugins":[
  #... other plugins
  {"name":"tsdb",
   "events":["alert","heartbeat"],
   "options":{"host":"<SignalFx Bridge IP Addr>","port":13321}
  }
  ]
...
```

# Configuration
The agent is configured by environment variables.  Configuration variables are:

 - `SIGNALFX_ACCESS_TOKEN` (**required**) - The SignalFx access token for the org
	 you want to receive the metrics

 - `CLOUDFOUNDRY_API_URL` (**required**) - The URL (including scheme) of the CF
	 Cloud Controller API.  This URL must be accessible by this app.

 - `CF_USERNAME` (**required**) - The client username for this app to access
	 the CF API and Firehose.  Admin (read-only) access is required for the CF
	 API.

 - `CF_PASSWORD` (**required**) - The client secret for the above user

 - `CF_UAA_URL` (**required**) - The URL of the CF UAA server (including scheme and
	 port).

 - `BOSH_DIRECTOR_URL` (**required**) - The URL (including scheme and port) of
	 the BOSH Director API

 - `BOSH_CLIENT_ID` (**required**) - The client username for this app to access
	 the BOSH Director API.

 - `BOSH_CLIENT_SECRET` (**required**) - The client secret for the above user

 - `TRAFFIC_CONTROLLER_URL` (optional) - The URL to the traffic controller.
	 This will be autodiscovered from the CF API if left blank

 - `FIREHOSE_SUBSCRIPTION_ID` (optional, default: *signalfx*) - The subscription id for
	 the Firehose nozzle.  This generally shouldn't need to be changed unless
	 running multiple deployments.

 - `FIREHOSE_IDLE_TIMEOUT_SECONDS` (optional, default: 20) - The number of
	 seconds to wait while the firehose is idle before timing out and
	 reconnecting.  The default is generally plenty of time but could be
	 shortened if connection drops are a frequent occurrance in your network
	 environment.

 - `FIREHOSE_RECONNECT_DELAY_SECONDS` (optional, default: 5) - The number of
	 seconds to wait before reconnecting to the Firehose after a timeout or
	 authentication token expiry.  The default should be fine for most all
	 cases, but could be shortened as well.

 - `DEPLOYMENTS_TO_INCLUDE` (optional, default: all) - A whitelist of BOSH
	 deployments to send metrics for.  If left blank (the default), all
	 deployments will be sent.  Separate multiple deployments with ";". Ex.
	 `cf-1;redis-1`.

 - `METRICS_TO_EXCLUDE` (optional, default: none) - A list of metrics to filter
	 out and not send to SignalFx.  This can be useful to keep your DPM lower.
	 There is [a list of Firehose
	 metrics](https://docs.cloudfoundry.org/running/all_metrics.html) to use
	 for reference.  Note that this app sends metrics with the origin name
	 prefixed, separated by a '.', so you will have to include that along with
	 the metric name in this value.  For example, if you didn't want to send
	 the number of active goroutines (the `numGoRoutines` metric) in the
	 *auctioneer* component, you would add the value `auctioneer.numGoRoutines`
	 to this config option.  Multiple metric names should be separated with
	 ';'.

 - `FLUSH_INTERVAL_SECONDS` (optional, default: 3) - How long to buffer metrics
	 before sending them to SignalFx.  A shorter time means quicker updates in
	 the dashboards, but if it is too short the overhead from so many HTTP
	 requests could cause metric buffers to continue to grow and be hard to
	 empty.

 - `INSECURE_SSL_SKIP_VERIFY` (optional, default: false) - Whether to skip TLS
	 cert verification.  This can be useful for testing environments.

 - `APP_METADATA_CACHE_EXPIRY_SECONDS` (optional, default: 300) - Each metrics
	 that comes off of the firehose about a CF app only contains an app GUID.
	 This means that in order to get human-readable information about an app
	 (e.g. name, space/org name) to send as dimensions to SignalFx, we need to
	 ask the CF API.  This settings determines how long the bridge caches the
	 app metadata before refetching it from the CF API.

 - `SIGNALFX_INGEST_URL` (optional) - You can change this if you are using the
	 MetricProxy to forward metrics.  Should be the full URL including the
	 datapoint path.


These values can be configured by the end user via the tile in Ops Manager
(Pivotal CF only) or in the deployment manifest for the BOSH release.

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

You can do the same with the `uaac` CLI tool.

## BOSH Director Access
The Bridge needs access to the BOSH directors to correctly link BOSH Health
Metrics to IP addresses of the VMs (unfortunately the BOSH HM Metrics don't
include IP address).  See [Creating a BOSH
Client](https://docs.pivotal.io/pivotalcf/1-10/customizing/opsmanager-create-bosh-client.html)
on how to do this with Pivotal Cloud Foundry.  If you are using a non-Pivotal
CF deployments, steps will differ but be mostly the same.

# Build

This app **requires Go 1.17+**.

```sh
$ make signalfx-bridge
$ # Set envvars (see above)
$ ./signalfx-bridge
```

# Tests

 The tests can be executed by:

```
go test -race ./...

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
