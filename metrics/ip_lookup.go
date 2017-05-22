package metrics

import (
	"sync"

    "github.com/cloudfoundry/sonde-go/events"
)

// This is a wrapper around a map to provide an IP Address cache.  The problem
// is that the BOSH HM metrics (fetched by tsdb_server.go) don't include IP
// address, but the Firehose metrics do.  On the Firehose, the `index` property
// corresponds to the `id` tag on the HM metrics.  So, in order to make BOSH
// metrics have the IP address, we can populate a shared instance of this class
// with values mapping from index/id to IP Address, which the tsdb_server
// module can use to properly populate the `host` dimension of the datapoints
// it sends to SignalFx.

type IPLookup struct {
	m         map[string]string
	lock      sync.RWMutex
	envelopes chan *events.Envelope
	stop      chan bool
}

func NewIPLookup() *IPLookup {
	return &IPLookup {
		m:         make(map[string]string),
		lock:      sync.RWMutex{},
		envelopes: make(chan *events.Envelope, 1000),
		stop:      make(chan bool),
	}
}

func (o *IPLookup) ListenForEnvelopes() {
	for {
		select {
		case <-o.stop:
			return
		case env := <-o.envelopes:
			if env.GetIp() != "" {
				o.addIPAddress(env.GetIndex(), env.GetIp())
			}
		}
	}
}

// Use a channel for envelopes so that we don't block threads trying to add to
// the map (it could block due to the write lock on the map).  The channel this
// sends to is buffered to hopefully avoid ever blocking.
func (o *IPLookup) SubmitEnvelope(env *events.Envelope) {
	o.envelopes <- env
}

func (o *IPLookup) Stop() {
	o.stop <- true
}

func (o *IPLookup) addIPAddress(bosh_id string, ip_addr string) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.m[bosh_id] = ip_addr
}

func (o *IPLookup) GetIPAddress(bosh_id string) string {
	o.lock.RLock()
	defer o.lock.RUnlock()

	return o.m[bosh_id]
}
