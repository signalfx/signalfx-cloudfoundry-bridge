package metrics_test

import (
    "os"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/signalfx/signalfx-bridge/metrics"
)

var _ = Describe("NozzleConfig", func() {
    BeforeEach(func() {
        os.Clearenv()
    })

    It("populates config values with environmental variables", func() {
        os.Setenv("CLOUDFOUNDRY_API_URL", "https://api.walnut-env.cf-app.com")
        os.Setenv("UAA_URL", "https://uaa.walnut-env.cf-app.com")
        os.Setenv("UAA_USERNAME", "env-user")
        os.Setenv("UAA_PASSWORD", "env-user-password")
        os.Setenv("TRAFFIC_CONTROLLER_URL", "wss://doppler.walnut-env.cf-app.com:4443")
        os.Setenv("FIREHOSE_SUBSCRIPTION_ID", "env-signalfx-nozzle")
        os.Setenv("FIREHOSE_IDLE_TIMEOUT_SECONDS", "60")
        os.Setenv("FIREHOSE_RECONNECT_DELAY_SECONDS", "5")
        os.Setenv("DEPLOYMENTS_TO_WATCH", "test1,test2")
        os.Setenv("FLUSH_INTERVAL_SECONDS", "25")
        os.Setenv("INSECURE_SSL_SKIP_VERIFY", "false")
        os.Setenv("SIGNALFX_INGEST_URL", "http://10.10.10.10")
        os.Setenv("SIGNALFX_ACCESS_TOKEN", "s3cr3t")


        conf, err := metrics.GetConfigFromEnv()
        Expect(err).ToNot(HaveOccurred())
        Expect(conf.CloudFoundryApiUrl).To(Equal("https://api.walnut-env.cf-app.com"))
        Expect(conf.UAAURL).To(Equal("https://uaa.walnut-env.cf-app.com"))
        Expect(conf.Username).To(Equal("env-user"))
        Expect(conf.Password).To(Equal("env-user-password"))
        Expect(conf.TrafficControllerURL).To(Equal("wss://doppler.walnut-env.cf-app.com:4443"))
        Expect(conf.FirehoseSubscriptionID).To(Equal("env-signalfx-nozzle"))
        Expect(conf.FirehoseIdleTimeoutSeconds).To(Equal(60))
        Expect(conf.FirehoseReconnectDelaySeconds).To(Equal(5))
        Expect(conf.DeploymentsToWatch).To(BeEquivalentTo([]string{"test1", "test2"}))
        Expect(conf.FlushIntervalSeconds).To(BeEquivalentTo(25))
        Expect(conf.InsecureSSLSkipVerify).To(Equal(false))
        Expect(conf.AppMetadataCacheExpirySeconds).To(Equal(300))
        Expect(conf.SignalFxIngestURL).To(Equal("http://10.10.10.10"))
        Expect(conf.SignalFxAccessToken).To(Equal("s3cr3t"))
    })
})
