package main

import (
	//"github.com/blevesearch/bleve"
        "github.com/gorilla/mux"
        "net/http"
        "fmt"
        "encoding/json"
	"io/ioutil"
        //"strings"
        //"time"
)


func (db *Datasets) NewAPI() http.Handler {
	router := mux.NewRouter()
        router.StrictSlash(true)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                sz, _:= db.index.DocCount()
                status := map[string]interface{} {
                        "size": int(sz),
                        "version": fmt.Sprintf("%s - %s", Name, Version),
                }
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(status)
        }).Methods("GET")


	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
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

	return router
}
