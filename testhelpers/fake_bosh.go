package testhelpers

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "regexp"
    "strconv"
    "strings"
)


// Provides mock data for the BoshClient


type FakeBosh struct {
    server     *httptest.Server
    // deployment name -> VM JSON objects
    vmsJson    map[string][]string
    // task id -> deploymentName
    tasks      map[string]string
    nextTaskId int
}

func NewFakeBosh() *FakeBosh {
    return &FakeBosh{
        vmsJson: make(map[string][]string),
        tasks:   make(map[string]string),
        nextTaskId: 1,
    }
}

func (f *FakeBosh) Start() {
    f.server = httptest.NewUnstartedServer(f)
    f.server.Start()
}

func (f *FakeBosh) Close() {
    f.server.Close()
}

func (f *FakeBosh) URL() string {
    return f.server.URL
}

func (f *FakeBosh) AddVM(deploymentName string, id string, ip string) {
    // The JSON has to be one line since the API expects you to split objects by \n
    vmJson := fmt.Sprintf(`{"vm_cid":"i-0a095c75f4a0a1285","disk_cid":null,"disk_cids":[],"ips":["%s"],"dns":[],"agent_id":"1b7458a3-2fb6-44df-af95-3b7a0201f644","job_name":"loggregator_trafficcontroller","index":1,"job_state":"running","state":"started","resource_pool":"t2.micro","vm_type":"t2.micro","vitals":{"cpu":{"sys":"0.2","user":"0.1","wait":"0.1"},"disk":{"ephemeral":{"inode_percent":"1","percent":"2"},"system":{"inode_percent":"30","percent":"35"}},"load":["0.00","0.00","0.00"],"mem":{"kb":"109548","percent":"11"},"swap":{"kb":"0","percent":"0"}},"processes":[{"name":"consul_agent","state":"running","uptime":{"secs":95006},"mem":{"kb":19768,"percent":1.9},"cpu":{"total":0}},{"name":"loggregator_trafficcontroller","state":"running","uptime":{"secs":95004},"mem":{"kb":15700,"percent":1.5},"cpu":{"total":0}},{"name":"metron_agent","state":"running","uptime":{"secs":95003},"mem":{"kb":12868,"percent":1.2},"cpu":{"total":0}},{"name":"route_registrar","state":"running","uptime":{"secs":95002},"mem":{"kb":18640,"percent":1.8},"cpu":{"total":0}}],"resurrection_paused":false,"az":"us-west-2a","id":"%s","bootstrap":false,"ignore":false}`, ip, id)

    f.vmsJson[deploymentName] = append(f.vmsJson[deploymentName], vmJson)
}

// Returns the app name as "app-<guid>" based on the guid passed in the path.
// This is meant to fake the `/v2/apps/<guid>` path, as well as the `/v2/info` path.
func (f *FakeBosh) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    vmsRegexp := regexp.MustCompile(`^/deployments/(.+)/vms$`)
    tasksRegexp := regexp.MustCompile(`^/tasks/(\d+)$`)
    tasksOutputRegexp := regexp.MustCompile(`^/tasks/(\d+)/output$`)

    if vmsRegexp.MatchString(r.URL.Path) {
        groups := vmsRegexp.FindStringSubmatch(r.URL.Path)
        deploymentName := groups[1]

        taskIdStr := strconv.Itoa(f.nextTaskId)
        f.tasks[taskIdStr] = deploymentName

        rw.Header().Set("Location", "/tasks/" + taskIdStr)
        rw.WriteHeader(302)

        f.nextTaskId++
    } else if tasksRegexp.MatchString(r.URL.Path) {
        // Just always make it done.  This could do a round of "processing" to
        // test the polling code more thoroughly.
        rw.Write([]byte(`{"state": "done"}`))
    } else if tasksOutputRegexp.MatchString(r.URL.Path) {
        groups := tasksOutputRegexp.FindStringSubmatch(r.URL.Path)
        taskId := groups[1]

        deploymentName := f.tasks[taskId]
        vmJson := f.vmsJson[deploymentName]
        rw.Write([]byte(strings.Join(vmJson, "\n")))
    }
}

