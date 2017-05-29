package metrics

import (
    "crypto/tls"
    "io/ioutil"
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "time"
    "log"
    //"github.com/davecgh/go-spew/spew"
)

type Deployment struct {
    Name string
}

type BoshVM struct {
    AgentID string `json:"agent_id"`
    VmCid   string `json:"vm_cid"`
    JobName string `json:"job_name"`
    Index    int
    Ips      []string `json:"ips"`
    Id       string
}

type BoshTask struct {
    State string
}

type BoshInfo struct {
    UserAuthentication struct {
        Type string
        Options struct {
            URL string
        }
    } `json:"user_authentication"`
}

type BoshClient struct {
    authTokenFetcher      AuthTokenFetcher
    baseURL               string
    client                *http.Client
    // Cached value from the authTokenFetcher
    authToken             string
    // Whether we're in the middle of retrying with a new auth token
    retrying              bool
    // How long to wait for the VM info task before canceling
    VMFetchTaskTimeoutSeconds int
}

// Makes an HTTP Client that times out in 10 seconds and doesn't follow
// redirects
func makeBoshHttpClient(skipSSLVerify bool) *http.Client {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerify},
        Proxy: http.ProxyFromEnvironment,
    }

    return &http.Client{
               Transport: tr,
               Timeout: time.Duration(10) * time.Second,
               // disabled redirect following
               CheckRedirect: func(req *http.Request, via []*http.Request) error {
                   return http.ErrUseLastResponse
               },
           }
}

func NewBoshClient(boshUrl string, authTokenFetcher AuthTokenFetcher, skipSSLVerify bool) *BoshClient {
    return &BoshClient {
        baseURL:          boshUrl,
        authTokenFetcher: authTokenFetcher,
        retrying:         false,
        client:           makeBoshHttpClient(skipSSLVerify),
        VMFetchTaskTimeoutSeconds:   60,
    }
}

func GetBoshUAAUrl(boshUrl string, skipSSLVerify bool) string {
    client := makeBoshHttpClient(skipSSLVerify)
    resp, err := client.Get(boshUrl + "/info")
    if err != nil {
        log.Fatal("Could not get BOSH /info endpoint: ", err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal("Could not read BOSH /info endpoint: ", err)
    }

    info := BoshInfo{}
    err = json.Unmarshal(body, &info)
    if err != nil {
        log.Fatal("Could not parse BOSH /info response: ", body, err)
    }

    if info.UserAuthentication.Type != "uaa" {
        log.Fatalf("This BOSH client only knows how to authenticate to BOSH"+
                   "using UAA and not %s", info.UserAuthentication.Type)
    }
    return info.UserAuthentication.Options.URL
}

func (o *BoshClient) NewGetRequest(path string) *http.Request {
    req, err := http.NewRequest("GET", o.baseURL + path, nil)
    if err != nil {
        log.Panic("Something is wrong with the BOSHClient config: ", err)
    }

    return req
}

// Returns the body text and response, or empty bytes/nil if there was error
func (o *BoshClient) doRequest(req *http.Request, expectedStatus int) ([]byte, *http.Response) {
    if o.authToken == "" {
        log.Printf("Fetching BOSH Auth Token")
        o.authToken = o.authTokenFetcher.FetchAuthToken()
    }

    // The uaa token includes the "bearer " prefix
    req.Header.Set("Authorization", o.authToken)

    resp, err := o.client.Do(req)
    if err != nil {
        log.Printf("Error fetching BOSH metadata (URL: %s): %v", req.URL.String(), err)
        return nil, nil
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        if !o.retrying {
            log.Print("BOSH UAA Token is not working, retrying with new token...")
            o.retrying = true
            o.authToken = ""
            return o.doRequest(req, expectedStatus)
        } else {
            log.Fatal("A new auth token didn't help authenticating to the BOSH " +
                      "Director.  Please check configuration.")
        }
    }
    o.retrying = false

    bodyText, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response from BOSH Director (URL: %s): %v", req.URL.String(), err)
        return nil, resp
    }

    if resp.StatusCode != expectedStatus {
        log.Printf("Unexpected status %d (expected %d) fetching BOSH metadata (URL: %s): %s",
                   resp.StatusCode, expectedStatus, req.URL.String(), bodyText)
        return nil, resp
    }

    return bodyText, resp
}

func (o *BoshClient) fetchDeployments() []Deployment {
    respText, _ := o.doRequest(o.NewGetRequest("/deployments"), 200)

    if len(respText) == 0 {
        // Errors would have been logged in the doRequest method
        return nil
    }

    deployments := make([]Deployment, 0)
    err := json.Unmarshal(respText, &deployments)
    if err != nil {
        log.Printf("Could not parse BOSH deployments response (%s): %s",
                   respText, err)
    }

    return deployments
}

// This is a long-running task so it's a bit more complex to handle.  The
// polling method seems simplest even if not very efficient.
func (o *BoshClient) fetchVMs(deploymentName string) []BoshVM {
    _, resp := o.doRequest(o.NewGetRequest("/deployments/" + deploymentName + "/vms?format=full"), 302)

    taskUrlStr := resp.Header.Get("Location")
    if len(taskUrlStr) == 0 {
        return nil
    }

    taskOutput := o.waitForTask(taskUrlStr)

    vms := make([]BoshVM, 0, 10)
    for _, vmJson := range strings.Split(string(taskOutput), "\n") {
        if vmJson == "" {
            continue
        }

        vm := BoshVM{}
        err := json.Unmarshal([]byte(vmJson), &vm)
        if err != nil {
            log.Printf("Could not parse deployment VM task output (%s): %s", vmJson, err)
            return nil
        }
        vms = append(vms, vm)
    }
    return vms
}

// taskUrlStr is like "https://bosh.dev/tasks/1234"
func (o *BoshClient) waitForTask(taskUrlStr string) []byte {
    waitStart := time.Now()
    for {
        taskUrl, err := url.Parse(taskUrlStr)
        if err != nil {
            log.Printf("Error parsing task url %s: %s", taskUrlStr, err)
            return nil
        }

        taskText, _ := o.doRequest(o.NewGetRequest(taskUrl.Path), 200)

        task := BoshTask{}
        err = json.Unmarshal(taskText, &task)
        if err != nil {
            log.Printf("Could not parse task JSON body (%s): %s", taskText, err)
            return nil
        }

        if task.State == "done" {
            outputText, _ := o.doRequest(o.NewGetRequest(taskUrl.Path + "/output?type=result"), 200)
            return outputText
        } else if waitStart.Add(time.Duration(o.VMFetchTaskTimeoutSeconds) * time.Second).Before(time.Now()) {
            log.Printf("Could not fetch VM stats from BOSH within %d seconds, try increasing timeout",
                       o.VMFetchTaskTimeoutSeconds)
            return nil
        } else {
            time.Sleep(500 * time.Millisecond)
        }
    }
}

