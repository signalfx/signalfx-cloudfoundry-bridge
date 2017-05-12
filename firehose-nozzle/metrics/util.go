package metrics

import (
    "bytes"
    "compress/gzip"
    "golang.org/x/net/context"
    "io/ioutil"
    "log"

    "github.com/signalfx/golib/datapoint"
)

type SignalFxClient interface {
    AddDatapoints(context.Context, []*datapoint.Datapoint) error
}


func Unzip(contents []byte) ([]byte, error) {
    buf := bytes.NewBuffer(contents)
    gzipReader, err := gzip.NewReader(buf)
    if err != nil {
        log.Printf("Fail to new gzip reader: %v", err)
        return nil, err
    }

    uncompressedData, err := ioutil.ReadAll(gzipReader)
    if err != nil {
        log.Printf("Fail to read content from gzip reader: %v", err)
        return nil, err
    }

    return uncompressedData, nil
}

func UnzipIgnoreError(contents []byte) []byte {
    uncompressedData, _ := Unzip(contents)

    return uncompressedData
}
