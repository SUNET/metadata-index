package main

import (
        "flag"
        //"os"
        "net/http"
)

var Name = "metadata-index"
var Version = "1.0.0"

var indexPath = flag.String("index", ".datasets", "index path")
var bindAddr = flag.String("bind", ":3000", "http listen address")
var batchSize = flag.Int("batchSize", 100, "batch size for indexing")
var serve = flag.Bool("serve", true, "start server (or only index and exit")
var rebuild = flag.Bool("rebuild", false, "force re-index")

func main() {
        flag.Parse()
        db := NewDatasets(*indexPath)
	db.Reload()
        http.Handle("/", db.NewAPI())
        log.Printf("Listening on %v", *bindAddr)
        log.Fatal(http.ListenAndServe(*bindAddr, nil))
}

