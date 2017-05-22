package metrics_test


import (
    "github.com/cloudfoundry-community/go-cfclient"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    . "github.com/signalfx/cloudfoundry-bridge/testhelpers"

	"github.com/signalfx/cloudfoundry-bridge/metrics"
)


var _ = Describe("AppMetadataFetcher", func() {
    var fakeCloudController *FakeCloudController
    var metadataFetcher *metrics.AppMetadataFetcher

    BeforeEach(func() {
        fakeCloudController = NewFakeCloudController()
        fakeCloudController.Start()

        cloudfoundryClient, err := cfclient.NewClient(&cfclient.Config{
            ApiAddress: fakeCloudController.URL(),
            Token: "testing",
            SkipSslValidation: true,
        })
        if err != nil {
            Fail("Could not setup CF client!")
        }

        metadataFetcher = metrics.NewAppMetadataFetcher(cloudfoundryClient)
    })

    AfterEach(func() {
        fakeCloudController.Close()
    })

    It("Caches data until expiry", func() {
        metadataFetcher.CacheExpirySeconds = 100
        guid := "1234-abcd"
        _ = metadataFetcher.GetAppNameForGUID(guid)
        _ = metadataFetcher.GetAppNameForGUID(guid)
        _ = metadataFetcher.GetAppNameForGUID(guid)
        Expect(fakeCloudController.AppReqCounts[guid]).To(Equal(1))
    })

    It("Refetches data after expiry", func() {
        metadataFetcher.CacheExpirySeconds = 0
        guid := "1234-abcd"
        _ = metadataFetcher.GetAppNameForGUID(guid)
        _ = metadataFetcher.GetAppNameForGUID(guid)
        _ = metadataFetcher.GetAppNameForGUID(guid)
        Expect(fakeCloudController.AppReqCounts[guid]).To(Equal(3))
    })
})
