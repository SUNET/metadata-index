package main

import (
        "flag"
        "net/http"
        "crypto/tls"
        "github.com/sunet/metadata-index/pkg/api"
        "github.com/sunet/metadata-index/pkg/datasets"
        "github.com/sunet/metadata-index/pkg/utils"
        "github.com/sirupsen/logrus"
)

var log = logrus.New()

var indexPath = flag.String("index", ".datasets", "index path")
var bindAddr = flag.String("bind", ":3000", "http listen address")
var batchSize = flag.Int("batchSize", 100, "batch size for indexing")
var serve = flag.Bool("serve", true, "start server (or only index and exit")
var rebuild = flag.Bool("rebuild", false, "force re-index")

func main() {
        flag.Parse()
        db := datasets.NewDatasets(*indexPath)
	db.Reload()

        http.Handle("/", api.NewAPI(&db))
        cfg := utils.GetTLSConfig()
        if cfg == nil {
                log.Fatal(http.ListenAndServe(*bindAddr, nil))
        } else {
                log.Debug("Enabling TLS...")
                listener, err := tls.Listen("tcp", *bindAddr, cfg)
                if err != nil {
                        log.Fatal(err)
                }
                log.Fatal(http.Serve(listener, nil))
        }
}

