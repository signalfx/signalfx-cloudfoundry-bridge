package metrics_test

import (
    "fmt"
    //"log"
    "strings"

    "github.com/cloudfoundry/sonde-go/events"
    "github.com/cloudfoundry-community/go-cfclient"
    "github.com/gogo/protobuf/proto"
    sfxproto "github.com/signalfx/com_signalfx_metrics_protobuf"
    "github.com/signalfx/golib/sfxclient"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    . "github.com/signalfx/signalfx-bridge/testhelpers"

    "github.com/signalfx/signalfx-bridge/metrics"
)


var _ = Describe("SignalFx Firehose Nozzle", func() {
    var fakeUAA *FakeUAA
    var fakeFirehose *FakeFirehose
    var fakeSignalFx *FakeSignalFx
    var fakeCloudController *FakeCloudController
    var config *metrics.Config
    var nozzle *metrics.SignalFxFirehoseNozzle
    var tokenFetcher *metrics.UAATokenFetcher
    var client *sfxclient.HTTPSink

    fakeFirehoseURL := func(ffh *FakeFirehose) string { return strings.Replace(ffh.URL(), "http:", "ws:", 1) }

    BeforeEach(func() {
        fakeUAA = NewFakeUAA("bearer", "123456789")
        fakeFirehose = NewFakeFirehose()
        fakeSignalFx = NewFakeSignalFx()
        fakeCloudController = NewFakeCloudController()

        fakeUAA.Start()
        fakeFirehose.Start()
        fakeSignalFx.Start()
        fakeCloudController.Start()

        tokenFetcher = &metrics.UAATokenFetcher{
            UaaUrl: fakeUAA.URL(),
        }

        config = &metrics.Config{
            CFUAAURL:             fakeUAA.URL(),
            TrafficControllerURL: fakeFirehoseURL(fakeFirehose),
            FlushIntervalSeconds: 1,
            FirehoseReconnectDelaySeconds: 1,
            FirehoseIdleTimeoutSeconds: 1,
            SignalFxIngestURL:    fakeSignalFx.URL(),
            SignalFxAccessToken:  "s3cr3t",
            DeploymentsToInclude: []string{"cf", "redis"},
            MetricsToExclude:     []string{"cc.log_count.all", "cc.log_count.debug"},
        }

        client = sfxclient.NewHTTPSink()
        client.DatapointEndpoint = fakeSignalFx.URL()

        cloudfoundryClient, err := cfclient.NewClient(&cfclient.Config{
            ApiAddress: fakeCloudController.URL(),
            Token: "testing",
            SkipSslValidation: true,
        })
        if err != nil {
            Fail("Could not setup CF client!")
        }
        metadataFetcher := metrics.NewAppMetadataFetcher(cloudfoundryClient)

        metricFilter := metrics.NewMetricFilter(config)

        fakeFirehose.KeepConnectionAlive()
        nozzle = metrics.NewSignalFxFirehoseNozzle(config, tokenFetcher, client, metadataFetcher, metricFilter)
    })

    AfterEach(func() {
        fakeUAA.Close()
        fakeFirehose.Close()
        fakeSignalFx.Close()
    })

    It("forwards ValueMetrics from the firehose", func(done Done) {
        defer close(done)
        defer GinkgoRecover()

        const ts int64 = 1000000000

        for i := 0; i < 10; i++ {
            envelope := events.Envelope{
                Origin:    proto.String("cc"),
                Timestamp: proto.Int64(ts),
                EventType: events.Envelope_ValueMetric.Enum(),
                ValueMetric: &events.ValueMetric{
                    Name:  proto.String(fmt.Sprintf("metricName-%d", i)),
                    Value: proto.Float64(float64(i)),
                    Unit:  proto.String("gauge"),
                },
                Deployment: proto.String("cf"),
                Job:        proto.String("doppler"),
                Index:      proto.String("abcdefg"),
                Ip:         proto.String("127.0.0.1"),
            }
            fakeFirehose.AddEvent(envelope)
        }

        go nozzle.Start()
        defer nozzle.Stop()

        By("Sending valid datapoints to the SignalFx ingest endpoint")
        datapoints := fakeSignalFx.GetIngestedDatapoints()

        By("Batching all the metrics in a short interval")
        Expect(datapoints).To(HaveLen(10))
        // Grab one out of the middle
        dp := datapoints[5]

        By("Converting ValueMetrics to Gauge values")
        Expect(dp.GetMetricType()).To(Equal(sfxproto.MetricType_GAUGE))

        By("Scaling timestamps from nanoseconds to milliseconds")
        Expect(dp.GetTimestamp()).To(Equal(ts/1000000))

        By("Forwarding the value unaltered")
        Expect(dp.GetValue().GetDoubleValue()).To(Equal(float64(5)))

        By("Prefixing the firehose origin field to the metric name")
        Expect(dp.GetMetric()).To(Equal("cc.metricName-5"))

        By("Setting the right dimensions")
        dimensions := ProtoDimensionsToMap(dp.GetDimensions())
        Expect(dimensions["metric_source"]).To(Equal("cloudfoundry"))
        Expect(dimensions["host"]).To(Equal("127.0.0.1"))
        Expect(dimensions["job"]).To(Equal("doppler"))
        Expect(dimensions["deployment"]).To(Equal("cf"))
        Expect(dimensions["bosh_id"]).To(Equal("abcdefg"))
    }, 5)

    It("forwards ContainerMetrics from the firehose", func(done Done) {
        defer close(done)
        defer GinkgoRecover()

        envelope := events.Envelope{
            Origin:    proto.String("rep"),
            Timestamp: proto.Int64(1000000000),
            EventType: events.Envelope_ContainerMetric.Enum(),
            ContainerMetric: &events.ContainerMetric{
                ApplicationId:  proto.String("testapp"),
                InstanceIndex: proto.Int32(2),
                CpuPercentage: proto.Float64(5.5),
                MemoryBytes: proto.Uint64(1000),
                DiskBytes: proto.Uint64(1000),
                MemoryBytesQuota: proto.Uint64(10000),
                DiskBytesQuota: proto.Uint64(10000),
            },
            Deployment: proto.String("cf"),
            Job:        proto.String("diego"),
            Index:      proto.String("abcdefg"),
            Ip:         proto.String("127.0.0.1"),
        }
        fakeFirehose.AddEvent(envelope)

        go nozzle.Start()
        defer nozzle.Stop()

        datapoints := fakeSignalFx.GetIngestedDatapoints()
        By("Splitting a single ContainerMetric to 5 datapoints")
        Expect(datapoints).To(HaveLen(5))

        var metricNames [5]string
        for i, dp := range datapoints {
            metricNames[i] = dp.GetMetric()
        }
        Expect(metricNames).To(ConsistOf(
            "container.cpu_percentage",
            "container.memory_bytes",
            "container.disk_bytes",
            "container.memory_bytes_quota",
            "container.disk_bytes_quota"))

        By("Setting the right dimensions")
        dimensions := ProtoDimensionsToMap(datapoints[0].GetDimensions())
        properties := ProtoPropertiesToMap(datapoints[0].GetProperties())

        Expect(dimensions["metric_source"]).To(Equal("cloudfoundry"))
        Expect(dimensions["host"]).To(Equal("127.0.0.1"))
        Expect(dimensions["job"]).To(Equal("diego"))
        Expect(dimensions["deployment"]).To(Equal("cf"))
        Expect(dimensions["bosh_id"]).To(Equal("abcdefg"))
        Expect(dimensions["app_id"]).To(Equal("testapp"))
        Expect(dimensions["app_instance_index"]).To(Equal("2"))
        Expect(properties["app_name"]).To(Equal("app-testapp"))
        Expect(properties["app_org"]).To(Equal("myorg"))
        Expect(properties["app_space"]).To(Equal("myspace"))
    }, 5)

    It("excludes metrics in blacklist", func(done Done) {
        defer close(done)
        defer GinkgoRecover()

        envelope := events.Envelope{
            Origin:    proto.String("cc"),
            Timestamp: proto.Int64(10000000000),
            EventType: events.Envelope_ValueMetric.Enum(),
            ValueMetric: &events.ValueMetric{
                Name:  proto.String(fmt.Sprintf("log_count.all")),
                Value: proto.Float64(1000),
                Unit:  proto.String("gague"),
            },
            Deployment: proto.String("cf"),
            Job:        proto.String("doppler"),
            Index:      proto.String("abcdefg"),
            Ip:         proto.String("127.0.0.1"),
        }
        fakeFirehose.AddEvent(envelope)

        go nozzle.Start()
        defer nozzle.Stop()

        fakeSignalFx.EnsureNoDatapoints()
    }, 5)

    Context("when the firehose sends an error", func() {
        It("should reconnect with different token", func(done Done) {
            defer close(done)
            defer GinkgoRecover()

            go nozzle.Start()
            defer nozzle.Stop()

            Eventually(fakeFirehose.Requested).Should(BeTrue())
            Expect(fakeFirehose.LastAuthorization()).To(Equal("bearer 123456789"))

            fakeUAA.SetAccessToken("abcdefghi")
            fakeFirehose.CloseAliveConnection()
            Eventually(fakeFirehose.LastAuthorization, 2).Should(Equal("bearer abcdefghi"))
        }, 3)
    })

    Context("when the firehose shuts down without notice", func() {
        It("should reconnect with different token", func(done Done) {
            defer close(done)
            defer GinkgoRecover()

            go nozzle.Start()
            defer nozzle.Stop()

            Eventually(fakeFirehose.Requested).Should(BeTrue())
            Expect(fakeFirehose.LastAuthorization()).To(Equal("bearer 123456789"))

            fakeUAA.SetAccessToken("abcdefghi")

            fakeFirehose.Close()
            fakeFirehose.Start()
            config.TrafficControllerURL = fakeFirehoseURL(fakeFirehose)

            Eventually(fakeFirehose.LastAuthorization, 3).Should(Equal("bearer abcdefghi"))
        }, 3)
    })

})
