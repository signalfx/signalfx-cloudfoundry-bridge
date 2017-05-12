# SignalFx Cloud Foundry Integration

There are two main compontents to our integration with Cloud Foundry (CF):

 - **SignalFx Firehose Nozzle**: This is technically both a [CF firehose
     nozzle](https://docs.cloudfoundry.org/loggregator/architecture.html) as
     well as a BOSH Health Monitor receiver.  The firehose nozzle picks up all
     of the metrics pertaining to the CF components as well as basic metrics on
     application containers (disk, memory, and cpu).  The BOSH HM receiver gets
     metrics about the VMs managed by the BOSH Director.  All of these metrics
     are forwarded to SignalFx.  See [the firehose
     README](firehose-nozzle/README.md) for more details.

 - **Firehose Nozzle BOSH Release**: This is the a BOSH release of our firehose
     nozzle.  This dir also contains logic to create a Pivotal CF Tile for use
     with the PCF Ops Manager.  The release can also be used in generic CF
     deployments.

 - **SignalFx Agent BOSH Release**: This is necessary to monitor your services
     that run in BOSH-managed VMs (i.e. not CF "applications").  This release
     can be added to your deployment manifest to cause our agent to be
     installed and configured on your machines.  See the
     [README](agent-bosh-release/README.md) for more details.
