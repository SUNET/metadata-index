package main

import (
	"github.com/blevesearch/bleve"
        "github.com/gorilla/mux"
        "net/http"
        "fmt"
        "encoding/json"
	"io/ioutil"
        "strings"
        //"time"
)


func (db *Datasets) NewAPI() http.Handler {
	router := mux.NewRouter()
        router.StrictSlash(true)

	router.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
                sz, _:= db.index.DocCount()
                status := map[string]interface{} {
                        "size": int(sz),
                        "version": fmt.Sprintf("%s - %s", Name, Version),
                }
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(status)
        }).Methods("GET")


	router.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		// Read body
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		url := string(b)
		var m manifest
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		b, err = ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(b, &m)
		if err != nil {
			log.Fatal(err)
		}
		db.index.Index(m.ID, m)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}).Methods("POST")

	router.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		qm := r.URL.Query()
                var searchResults *bleve.SearchResult
                var err error
                if val, ok := qm["q"]; ok {
                        q := strings.ToLower(val[0])
                        query := bleve.NewQueryStringQuery(q)
                        log.Printf("%s", query)
                        search := bleve.NewSearchRequest(query)
                        search.Size = 100
                        search.Fields = append(search.Fields, "data")
                        searchResults, err = db.index.Search(search)
                        if err != nil {
                                log.Fatal(err) // better to give up
                        }
                } else {
                        sz, _ :=  db.index.DocCount()
                        query := bleve.NewMatchAllQuery()
                        search := bleve.NewSearchRequest(query)
                        search.Size = int(sz)
                        searchResults, err = db.index.Search(search)
                        if err != nil {
                                log.Fatal(err) // better to give up
                        }
                }
                w.Header().Set("Content-Type", "application/json")
                enc := json.NewEncoder(w)
                comma := false
                w.Write([]byte("["))
                for _, hit := range searchResults.Hits {
                        if comma {
                                w.Write([]byte(","))
                        }
                        comma = true
                        enc.Encode(hit.ID)
                }
                w.Write([]byte("]"))
        }).Methods("GET")

	return router
}
