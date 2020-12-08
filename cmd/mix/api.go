package main

import (
	"github.com/blevesearch/bleve"
        "github.com/blevesearch/bleve/document"
        "github.com/gorilla/mux"
        "net/http"
        "reflect"
        "fmt"
        "encoding/json"
	"io/ioutil"
        "strings"
        //"time"
)

func fieldByTag(s reflect.Value, tag string) reflect.Value {
   //var mf ManifestInfo = ManifestInfo{}
   //st := reflect.TypeOf(mf)
   se := s.Elem()
   st := se.Type()
   for i := 0; i < st.NumField(); i++ {
      jsonTag, _ := st.Field(i).Tag.Lookup("json")
      if (jsonTag == tag) {
         return se.Field(i)
      }
   }
   return reflect.ValueOf(nil)
}

func doc2manifest(doc *document.Document, m *ManifestInfo) {
   mv := reflect.ValueOf(m)
   for _,field := range doc.Fields {
      fn := field.Name()
      fv := field.Value()
      f := fieldByTag(mv, fn)
      if (strings.HasPrefix(fn, "mm:manifest.")) {
         s := fn[12:]
         if (m.Manifest == nil) {
            m.Manifest = make([]TypedSchema, 0)
         }
         pos := field.ArrayPositions()
         i := int(pos[0])
         if (i >= len(m.Manifest)) {
            var tf TypedSchema
            m.Manifest = append(m.Manifest, tf)
         }
         sf := fieldByTag(reflect.ValueOf(&m.Manifest[i]), s)
         if (len(fv) > 0) {
            sf.SetString(string(fv))
         }
      } else {
         f.SetString(string(fv))
      }
   }
}

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
		var m ManifestInfo
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		b, err = ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(b, &m)
                m.ID = url //XXX maybe this is right?
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
                w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Methods", "GET");
                enc := json.NewEncoder(w)
                comma := false
                w.Write([]byte("["))
                for _, hit := range searchResults.Hits {
                        var m ManifestInfo
                        doc, _ := db.index.Document(hit.ID)
                        if comma {
                                w.Write([]byte(","))
                        }
                        comma = true
                        doc2manifest(doc, &m)
                        enc.Encode(m)
                }
                w.Write([]byte("]"))
        }).Methods("GET")

	return router
}
