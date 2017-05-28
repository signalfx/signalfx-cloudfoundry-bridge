// +build !race

package metrics_test


import (
    "fmt"
    "net"
    "strconv"
    "time"

    sfxproto "github.com/signalfx/com_signalfx_metrics_protobuf"
    "github.com/signalfx/golib/sfxclient"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    . "github.com/signalfx/signalfx-bridge/testhelpers"

    "github.com/signalfx/signalfx-bridge/metrics"
)


var _ = Describe("TSDBServer", func() {
    var fakeSignalFx *FakeSignalFx
    var fakeUAA *FakeUAA
    var fakeBosh *FakeBosh
    var sfxClient *sfxclient.HTTPSink
    var tsdbServer *metrics.TSDBServer
    var port int

    BeforeEach(func() {
        fakeUAA = NewFakeUAA("bearer", "123456789")
        fakeUAA.Start()

        fakeSignalFx = NewFakeSignalFx()
        fakeSignalFx.Start()

        fakeBosh = NewFakeBosh()
        fakeBosh.Start()

        sfxClient = sfxclient.NewHTTPSink()
        sfxClient.DatapointEndpoint = fakeSignalFx.URL()

        tokenFetcher := &metrics.UAATokenFetcher{
            UaaUrl: fakeUAA.URL(),
        }

        boshClient := metrics.NewBoshClient(fakeBosh.URL(), tokenFetcher, true)
        bosh := metrics.NewBoshMetadataFetcher(boshClient)

        port = 13321

        go func() {
            for {
                tsdbServer = metrics.NewTSDBServer(sfxClient, 1, port, bosh)
                err := tsdbServer.Start()
                if err != nil {
                    // Make the tests more robust by not being dependent on a
                    // single hard coded port
                    port += 1
                } else { break }
            }
        }()

        // Since TSDBServer.Start blocks if it binds successfully to the port,
        // we need to poll the port var until it stops changing.  This is still
        // theoretically subject to race conditions if the StartTSDBServer
        // method has started down a successful path but hasn't fully
        // configured the socket handler, which could cause sends in the test
        // to fail if they happen too fast.  If this proves to cause test
        // fragility, the best solution is to rework the code that sets up the
        // tcp socket to send a confirmation on a channel after binding to the
        // port but before going into the blocking listen loop.
        for {
            prevPort := port
            // Hopefully half a second is enough time to fail and try another
            // port
            time.Sleep(500 * time.Millisecond)
            if prevPort == port { break }
        }
    })

    AfterEach(func() {
        fakeSignalFx.Close()
        fakeUAA.Close()
        fakeBosh.Close()
        tsdbServer.Stop()
    })

    sendTSDBLine := func(line string) {
        conn, err := net.Dial("tcp", "localhost:" + strconv.Itoa(port))
        if err != nil {
            Fail(fmt.Sprint("Could not send to TSDBServer: ", err.Error()))
        }

        fmt.Fprintf(conn, line + "\n")
    }

    It("forwards VM metrics to SignalFx", func() {
        sendTSDBLine("put system.disk.ephemeral.percent 1493049192 2 deployment=p-metrics-d9889b7d6988533733d6 id=84d86321-8040-464f-be37-2389135e16bc index=0 job=opentsdb-metrics role=unknown")
        sendTSDBLine("put system.cpu.user 1493049198 0.6 deployment=cf-1f83d62c70fa873ce366 id=cd14da4b-b764-4e45-b6c3-142a8a058f4a index=0 job=consul_server role=unknown")

        datapoints := fakeSignalFx.GetIngestedDatapoints()
        Expect(datapoints).To(HaveLen(2))

        dp := datapoints[1]
        Expect(dp.GetMetric()).To(Equal("system.cpu.user"))

        By("Sending them to SignalFx as gagues")
        Expect(dp.GetMetricType()).To(Equal(sfxproto.MetricType_GAUGE))

        By("Scaling the timestamp from seconds to milliseconds")
        Expect(dp.GetTimestamp()).To(Equal(int64(1493049198000)))

        By("Setting the right dimensions on the datapoint")
        dimensions := ProtoDimensionsToMap(dp.GetDimensions())
        Expect(dimensions["metric_source"]).To(Equal("cloudfoundry"))
        Expect(dimensions["deployment"]).To(Equal("cf-1f83d62c70fa873ce366"))
        Expect(dimensions["id"]).To(Equal("cd14da4b-b764-4e45-b6c3-142a8a058f4a"))
        Expect(dimensions["job"]).To(Equal("consul_server"))
        Expect(dimensions["index"]).To(Equal("0"))
        Expect(dimensions["role"]).To(Equal("unknown"))

        By("Passing through the value unaltered")
        Expect(dp.GetValue().GetDoubleValue()).To(Equal(0.6))
    })

    It("uses the BOSH metadata fetcher to add host dimension", func() {
        fakeBosh.AddVM("p-metrics-d9889b7d6988533733d6", "84d86321-8040-464f-be37-2389135e16bc", "10.0.5.5")
        fakeBosh.AddVM("cf-1f83d62c70fa873ce366", "cd14da4b-b764-4e45-b6c3-142a8a058f4a", "10.0.10.10")

        sendTSDBLine("put system.disk.ephemeral.percent 1493049192 2 deployment=p-metrics-d9889b7d6988533733d6 id=84d86321-8040-464f-be37-2389135e16bc index=0 job=opentsdb-metrics role=unknown")
        sendTSDBLine("put system.cpu.user 1493049198 0.6 deployment=cf-1f83d62c70fa873ce366 id=cd14da4b-b764-4e45-b6c3-142a8a058f4a index=0 job=consul_server role=unknown")

        datapoints := fakeSignalFx.GetIngestedDatapoints()
        Expect(datapoints).To(HaveLen(2))

        dp := datapoints[0]
        dimensions := ProtoDimensionsToMap(dp.GetDimensions())
        Expect(dimensions["host"]).To(Equal("10.0.5.5"))

        dp = datapoints[1]
        dimensions = ProtoDimensionsToMap(dp.GetDimensions())
        Expect(dimensions["host"]).To(Equal("10.0.10.10"))
    })
})
