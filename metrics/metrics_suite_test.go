package metrics_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "testing"
)

func TestSignalFXfirehosenozzle(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "SignalFx Metrics Suite")
}
