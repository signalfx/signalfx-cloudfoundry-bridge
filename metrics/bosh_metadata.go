package metrics

import (
    "log"
    "time"
    //"github.com/davecgh/go-spew/spew"
)


type BoshMetadataFetcher struct {
    client              *BoshClient
    vmCache             map[string]*BoshVM
    // deployment name -> time last updated
    vmCacheLastUpdate   map[string]time.Time
    CacheExpirySeconds  int
    // Golang's version of a set
    vmsNotFound         map[string]bool
}

const defaultBoshCacheExpirySeconds = 180

func NewBoshMetadataFetcher(boshClient *BoshClient) *BoshMetadataFetcher {
    return &BoshMetadataFetcher{
        client:              boshClient,
        vmCache:             make(map[string]*BoshVM),
        vmCacheLastUpdate:   make(map[string]time.Time),
        CacheExpirySeconds:  defaultBoshCacheExpirySeconds,
        vmsNotFound:         make(map[string]bool),
    }
}

// This just returns the first ip address given by BOSH
func (o *BoshMetadataFetcher) GetVMIPAddress(deploymentName, vmId string) string {
    vm := o.getVM(deploymentName, vmId)
    if vm == nil || len(vm.Ips) == 0 {
        return ""
    }

    return vm.Ips[0]
}

func (o *BoshMetadataFetcher) refreshVMCacheFor(deploymentName string) {
    log.Print("Refreshing BOSH VM cache for ", deploymentName)
    vms := o.client.fetchVMs(deploymentName)
    // Use index since the value will be a copy of the pointed value and not
    // the value via the original pointer
    for i := range vms {
        o.vmCache[vms[i].Id] = &vms[i]
    }
    o.vmCacheLastUpdate[deploymentName] = time.Now()
}

func (o *BoshMetadataFetcher) getVM(deploymentName, vmId string) *BoshVM {
    lastUpdate := o.vmCacheLastUpdate[deploymentName]
    if lastUpdate.IsZero() {
        o.refreshVMCacheFor(deploymentName)
    } else {
        expiryTime := lastUpdate.Add(time.Duration(o.CacheExpirySeconds) * time.Second)
        if expiryTime.Before(time.Now()) {
            log.Print("Expiring BOSH vm cache for deployment: ", deploymentName)
            o.refreshVMCacheFor(deploymentName)
        }
    }

    vm, ok := o.vmCache[vmId]
    if !ok {
        // Make sure we don't get into infinite recursion by only refetching
        // for a vm once
        if o.vmsNotFound[vmId] {
            log.Printf("VM '%s' not found again, something is inconsistent.", vmId)
        } else {
            log.Printf("VM '%s' not found in cache, refetching...", vmId)
            o.vmsNotFound[vmId] = true

            o.refreshVMCacheFor(deploymentName)
            return o.getVM(deploymentName, vmId)
        }
    } else {
        delete(o.vmsNotFound, vmId)
    }

    return vm
}
