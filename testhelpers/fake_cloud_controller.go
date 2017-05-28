package testhelpers

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "regexp"
)


// Provides mock data for the Cloud Controller API Client


// Data for the /v2/info endpoint
const infoJSON = `
{
   "name": "cftest",
   "build": "a7463f020805516a312ba23414bcb0adcf73d455",
   "support": "support@example.com",
   "version": 0,
   "description": "",
   "authorization_endpoint": "https://login.local.pcfdev.io",
   "token_endpoint": "https://uaa.local.pcfdev.io",
   "min_cli_version": null,
   "min_recommended_cli_version": null,
   "api_version": "2.75.0",
   "app_ssh_endpoint": "ssh.local.pcfdev.io:2222",
   "app_ssh_host_key_fingerprint": "a6:d1:08:0b:b0:cb:9b:5f:c4:ba:44:2a:97:26:19:8a",
   "app_ssh_oauth_client": "ssh-proxy",
   "routing_endpoint": "https://api.local.pcfdev.io/routing",
   "logging_endpoint": "wss://loggregator.local.pcfdev.io:443",
   "doppler_logging_endpoint": "wss://doppler.local.pcfdev.io:443"
}
`

type FakeCloudController struct {
    server     *httptest.Server
    AppReqCounts  map[string]int
}

func NewFakeCloudController() *FakeCloudController {
    return &FakeCloudController{
        AppReqCounts: make(map[string]int),
    }
}

func (f *FakeCloudController) Start() {
    f.server = httptest.NewUnstartedServer(f)
    f.server.Start()
}

func (f *FakeCloudController) Close() {
    f.server.Close()
}

func (f *FakeCloudController) URL() string {
    return f.server.URL
}

// Returns the app name as "app-<guid>" based on the guid passed in the path.
// This is meant to fake the `/v2/apps/<guid>` path, as well as the `/v2/info` path.
func (f *FakeCloudController) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    appPathRegexp := regexp.MustCompile(`^/v2/apps/([\w-]+)$`)

    if r.URL.Path == "/v2/info" {
        rw.Write([]byte(infoJSON))
    } else if appPathRegexp.MatchString(r.URL.Path) {
        groups := appPathRegexp.FindStringSubmatch(r.URL.Path)
        guid := groups[1]

        f.AppReqCounts[guid] += 1

        // Extremely stripped down version of what CC actually returns
        rw.Write([]byte(fmt.Sprintf(`
            {
              "metadata": {
                "guid": "%s"
              },
              "entity": {
                "name": "app-%s",
                "space_guid": "25afdd92-2acc-49f7-9d5b-4206af993286",
                "space": {
                  "metadata": {
                    "guid": "25afdd92-2acc-49f7-9d5b-4206af993286"
                  },
                  "entity": {
                    "name": "myspace",
                    "organization_guid": "0175d8fd-e76b-4ca6-914d-7b3d2a4536b2",
                    "organization": {
                      "metadata": {
                        "guid": "0175d8fd-e76b-4ca6-914d-7b3d2a4536b2"
                      },
                      "entity": {
                        "name": "myorg"
                      }
                    }
                  }
                }
              }
            }
        `, guid, guid)))
    }
}

