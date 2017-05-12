package metrics

import (
    "log"
    "time"

    "github.com/cloudfoundry-community/go-cfclient"
)

// The AppMetadataFetcher is used to get the human names for applications to
// send as a dimension, as well as the org and space names.  The data coming
// out of the Firehose only includes an app GUID.

// The data is cached for `CacheExpirySeconds`, after which it is refetched
// from the CF API.  The CF API requires the scope/authority of
// `cloud_controller.admin_read_only` to pull this information.

type CacheEntry struct {
    App *cfclient.App
    InsertTime time.Time
}

type AppMetadataFetcher struct {
    appCache            map[string]*CacheEntry
    client              *cfclient.Client
    CacheExpirySeconds  int
}

const defaultCacheExpirySeconds = 5 * 60


func NewAppMetadataFetcher(client *cfclient.Client) *AppMetadataFetcher {
    return &AppMetadataFetcher{
        appCache: make(map[string]*CacheEntry),
        client: client,
        CacheExpirySeconds: defaultCacheExpirySeconds,
    }
}

func (a *AppMetadataFetcher) fetchApp(guid string) (*cfclient.App, error) {
    cacheEntry := a.appCache[guid]
    if cacheEntry != nil {
        expiryTime := cacheEntry.InsertTime.Add(time.Duration(a.CacheExpirySeconds) * time.Second)
        if expiryTime.Before(time.Now()) {
            log.Print("Expiring app metadata cache for ", guid)
            delete(a.appCache, guid)
            return a.fetchApp(guid)
        }
    } else {
        log.Print("Fetching app metadata for ", guid)
        app, err := a.client.AppByGuid(guid)
        if err != nil {
            log.Printf("Error fetching app %s: %v", guid, err)
            return nil, err
        }

        cacheEntry = &CacheEntry{&app, time.Now()}
        a.appCache[guid] = cacheEntry
    }

    return cacheEntry.App, nil
}

func (a *AppMetadataFetcher) GetAppNameForGUID(guid string) string {
    app, err := a.fetchApp(guid)
    if err != nil {
        return ""
    }

    return app.Name
}

func (a *AppMetadataFetcher) GetSpaceNameForGUID(guid string) string {
    app, err := a.fetchApp(guid)
    if err != nil {
        return ""
    }

    return app.SpaceData.Entity.Name
}

func (a *AppMetadataFetcher) GetOrgNameForGUID(guid string) string {
    app, err := a.fetchApp(guid)
    if err != nil {
        return ""
    }

    return app.SpaceData.Entity.OrgData.Entity.Name
}
