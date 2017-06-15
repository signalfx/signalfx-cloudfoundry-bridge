package metrics_test

import (
    "os"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/signalfx/signalfx-cloudfoundry-bridge/metrics"
)

var _ = Describe("NozzleConfig", func() {
    BeforeEach(func() {
        os.Clearenv()
    })

    It("populates config values with environmental variables", func() {
        os.Setenv("CLOUDFOUNDRY_API_URL", "https://api.walnut-env.cf-app.com")
        os.Setenv("CF_UAA_URL", "https://uaa.walnut-env.cf-app.com")
        os.Setenv("CF_USERNAME", "env-user")
        os.Setenv("CF_PASSWORD", "env-user-password")
        os.Setenv("BOSH_DIRECTOR_URL", "https://123.123.123.123:25555")
        os.Setenv("BOSH_CLIENT_ID", "bosh-username")
        os.Setenv("BOSH_CLIENT_SECRET", "bosh-password")
        os.Setenv("TRAFFIC_CONTROLLER_URL", "wss://doppler.walnut-env.cf-app.com:4443")
        os.Setenv("FIREHOSE_SUBSCRIPTION_ID", "env-signalfx-nozzle")
        os.Setenv("FIREHOSE_IDLE_TIMEOUT_SECONDS", "60")
        os.Setenv("FIREHOSE_RECONNECT_DELAY_SECONDS", "5")
        os.Setenv("DEPLOYMENTS_TO_INCLUDE", "test1;test2")
        os.Setenv("METRICS_TO_EXCLUDE", "cpu.idle; memory")
        os.Setenv("FLUSH_INTERVAL_SECONDS", "25")
        os.Setenv("INSECURE_SSL_SKIP_VERIFY", "false")
        os.Setenv("SIGNALFX_INGEST_URL", "http://10.10.10.10")
        os.Setenv("SIGNALFX_ACCESS_TOKEN", "s3cr3t")


        conf, err := metrics.GetConfigFromEnv()
        Expect(err).ToNot(HaveOccurred())
        Expect(conf.CloudFoundryApiURL).To(Equal("https://api.walnut-env.cf-app.com"))
        Expect(conf.CFUAAURL).To(Equal("https://uaa.walnut-env.cf-app.com"))
        Expect(conf.CFUsername).To(Equal("env-user"))
        Expect(conf.CFPassword).To(Equal("env-user-password"))
        Expect(conf.BoshDirectorURL).To(Equal("https://123.123.123.123:25555"))
        Expect(conf.BoshUsername).To(Equal("bosh-username"))
        Expect(conf.BoshPassword).To(Equal("bosh-password"))
        Expect(conf.TrafficControllerURL).To(Equal("wss://doppler.walnut-env.cf-app.com:4443"))
        Expect(conf.FirehoseSubscriptionID).To(Equal("env-signalfx-nozzle"))
        Expect(conf.FirehoseIdleTimeoutSeconds).To(Equal(60))
        Expect(conf.FirehoseReconnectDelaySeconds).To(Equal(5))
        Expect(conf.DeploymentsToInclude).To(BeEquivalentTo([]string{"test1", "test2"}))
        Expect(conf.MetricsToExclude).To(BeEquivalentTo([]string{"cpu.idle", "memory"}))
        Expect(conf.FlushIntervalSeconds).To(BeEquivalentTo(25))
        Expect(conf.InsecureSSLSkipVerify).To(Equal(false))
        Expect(conf.AppMetadataCacheExpirySeconds).To(Equal(300))
        Expect(conf.SignalFxIngestURL).To(Equal("http://10.10.10.10"))
        Expect(conf.SignalFxAccessToken).To(Equal("s3cr3t"))
    })
})
