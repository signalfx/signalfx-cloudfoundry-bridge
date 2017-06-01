package testhelpers

import (
    sfxproto "github.com/signalfx/com_signalfx_metrics_protobuf"
)


func ProtoDimensionsToMap(dims []*sfxproto.Dimension) (m map[string]string) {
    m = make(map[string]string)

    for _, d := range dims {
        m[d.GetKey()] = d.GetValue()
    }
    return m
}

func ProtoPropertiesToMap(props []*sfxproto.Property) (m map[string]string) {
    m = make(map[string]string)

    for _, p := range props {
		// We only use string props in this app
        m[p.GetKey()] = p.GetValue().GetStrValue()
    }
    return m
}


