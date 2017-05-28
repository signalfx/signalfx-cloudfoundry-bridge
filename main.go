package main

import (
    "errors"
    "log"
    "os"
    "os/signal"
    "runtime"
    "runtime/pprof"
    "syscall"

    "github.com/signalfx/golib/sfxclient"
    "github.com/cloudfoundry-community/go-cfclient"

    . "github.com/signalfx/signalfx-bridge/metrics"
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

    cfTokenFetcher := &UAATokenFetcher{
        UaaUrl:        config.CFUAAURL,
        Username:      config.CFUsername,
        Password:      config.CFPassword,
        SSLSkipVerify: config.InsecureSSLSkipVerify,
    }

    cloudfoundry, err := cfclient.NewClient(&cfclient.Config{
        ApiAddress: config.CloudFoundryApiURL,
        ClientID: config.CFUsername,
        ClientSecret: config.CFPassword,
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

    boshUAAUrl := GetBoshUAAUrl(config.BoshDirectorURL, config.InsecureSSLSkipVerify)
    boshTokenFetcher := &UAATokenFetcher{
        UaaUrl:        boshUAAUrl,
        Username:      config.BoshUsername,
        Password:      config.BoshPassword,
        SSLSkipVerify: config.InsecureSSLSkipVerify,
    }
    boshClient := NewBoshClient(config.BoshDirectorURL,
                                boshTokenFetcher,
                                config.InsecureSSLSkipVerify)
    bosh := NewBoshMetadataFetcher(boshClient)

    errChan := make(chan error)

    go func() {
        metadataFetcher := NewAppMetadataFetcher(cloudfoundry)
        metadataFetcher.CacheExpirySeconds = config.AppMetadataCacheExpirySeconds

        nozzle := NewSignalFxFirehoseNozzle(config, cfTokenFetcher, sfxClient, metadataFetcher)
        nozzle.Start()
        errChan <- errors.New("Firehose Nozzle quit unexpectedly")
    }()

    go func() {
        tsdbErr := NewTSDBServer(sfxClient, config.FlushIntervalSeconds, 0, bosh).Start()

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
