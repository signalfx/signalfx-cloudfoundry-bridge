package metrics

import (
    "github.com/caarlos0/env"
)

type Config struct {
    CloudFoundryApiUrl             string   `env:"CLOUDFOUNDRY_API_URL,required"`
    UAAURL                         string   `env:"UAA_URL,required"`
    Username                       string   `env:"UAA_USERNAME,required"`
    Password                       string   `env:"UAA_PASSWORD,required"`
    InsecureSSLSkipVerify          bool     `env:"INSECURE_SSL_SKIP_VERIFY" envDefault:"false"`

    // This will be populated automatically in the main package if not supplied
    TrafficControllerURL           string   `env:"TRAFFIC_CONTROLLER_URL" envDefault:""`
    FirehoseSubscriptionID         string   `env:"FIREHOSE_SUBSCRIPTION_ID" envDefault:"signalfx"`
    FlushIntervalSeconds           int      `env:"FLUSH_INTERVAL_SECONDS" envDefault:"3"`
    FirehoseIdleTimeoutSeconds     int      `env:"FIREHOSE_IDLE_TIMEOUT_SECONDS" envDefault:"20"`
    FirehoseReconnectDelaySeconds  int      `env:"FIREHOSE_RECONNECT_DELAY_SECONDS" envDefault:"5"`
	DeploymentsToWatch              []string `env:"DEPLOYMENTS_TO_WATCH" envDefault:"" envSeparator:","`

    AppMetadataCacheExpirySeconds  int       `env:"APP_METADATA_CACHE_EXPIRY_SECONDS" envDefault:"300"`

    SignalFxIngestURL              string   `env:"SIGNALFX_INGEST_URL"`
    SignalFxAccessToken            string   `env:"SIGNALFX_ACCESS_TOKEN,required"`
}

func GetConfigFromEnv() (*Config, error) {
    cfg := Config{}
    err := env.Parse(&cfg)

    return &cfg, err
}

