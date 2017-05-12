package main

import (
    "errors"
    "log"
    "os"
    "os/signal"
    "runtime"
    "runtime/pprof"
    "strings"
    "syscall"

    "github.com/signalfx/golib/sfxclient"
    "github.com/cloudfoundry-community/go-cfclient"

    . "github.com/signalfx/cloudfoundry-integration/firehose-nozzle/metrics"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    config, err := GetConfigFromEnv()
    if err != nil {
        log.Fatalf("Error in config: %s", err)
    }

    log.Printf("Using configuration values: %#v", config)

    threadDumpChan := registerGoRoutineDumpSignalChannel()
    defer close(threadDumpChan)
    go dumpGoRoutine(threadDumpChan)

    tokenFetcher := &UAATokenFetcher{
        UaaUrl:                config.UAAURL,
        Username:              config.Username,
        Password:              config.Password,
        InsecureSSLSkipVerify: config.InsecureSSLSkipVerify,
    }

    // The CF client lib doesn't like the token type prefixed to the token
    token := strings.Replace(tokenFetcher.FetchAuthToken(), "bearer ", "", 1)
    cloudfoundry, err := cfclient.NewClient(&cfclient.Config{
        ApiAddress: config.CloudFoundryApiUrl,
        Token: token,
        SkipSslValidation: config.InsecureSSLSkipVerify,
    })
    if err != nil {
        log.Fatal("Error initializing with the Cloud Foundry API: ", err)
    }

    if config.TrafficControllerURL == "" {
        config.TrafficControllerURL = cloudfoundry.Endpoint.DopplerEndpoint
    }


    sfxClient := sfxclient.NewHTTPSink()
    sfxClient.AuthToken = config.SignalFxAccessToken
    if config.SignalFxIngestURL != "" {
        sfxClient.DatapointEndpoint = config.SignalFxIngestURL
    }

    errChan := make(chan error)

    go func() {
        metadataFetcher := NewAppMetadataFetcher(cloudfoundry)
        metadataFetcher.CacheExpirySeconds = config.AppMetadataCacheExpirySeconds

        nozzle := NewSignalFxFirehoseNozzle(config, tokenFetcher, sfxClient, metadataFetcher)
        nozzle.Start()
        errChan <- errors.New("Firehose Nozzle quit unexpectedly")
    }()

    go func() {
        tsdbErr := StartTSDBServer(sfxClient, config.FlushIntervalSeconds, 0)
        errChan <- tsdbErr
    }()

    err = <-errChan
    log.Fatal(err)
}

func registerGoRoutineDumpSignalChannel() chan os.Signal {
    threadDumpChan := make(chan os.Signal, 1)
    signal.Notify(threadDumpChan, syscall.SIGUSR1)

    return threadDumpChan
}

func dumpGoRoutine(dumpChan chan os.Signal) {
    for range dumpChan {
        goRoutineProfiles := pprof.Lookup("goroutine")
        if goRoutineProfiles != nil {
            goRoutineProfiles.WriteTo(os.Stdout, 2)
        }
    }
}
