package api

import (
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/sunet/metadata-index/docs"
	"github.com/sunet/metadata-index/pkg/datasets"
	"github.com/sunet/metadata-index/pkg/manifest"
	"github.com/sunet/metadata-index/pkg/meta"
	"github.com/sunet/metadata-index/pkg/utils"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger"
)

var log = logrus.New()

func fieldByTag(s reflect.Value, tag string) reflect.Value {
	se := s.Elem()
	st := se.Type()
	for i := 0; i < st.NumField(); i++ {
		jsonTag, _ := st.Field(i).Tag.Lookup("json")
		if jsonTag == tag {
			return se.Field(i)
		}
	}
	return reflect.ValueOf(nil)
}

func doc2manifest(doc *document.Document, m *manifest.ManifestInfo) {
	mv := reflect.ValueOf(m)
	for _, field := range doc.Fields {
		fn := field.Name()
		fv := field.Value()
		f := fieldByTag(mv, fn)
		if strings.HasPrefix(fn, "mm:manifest.") {
			s := fn[12:]
			if m.Manifest == nil {
				m.Manifest = make([]manifest.TypedSchema, 0)
			}
			pos := field.ArrayPositions()
			i := int(pos[0])
			if i >= len(m.Manifest) {
				var tf manifest.TypedSchema
				m.Manifest = append(m.Manifest, tf)
			}
			sf := fieldByTag(reflect.ValueOf(&m.Manifest[i]), s)
			if len(fv) > 0 {
				sf.SetString(string(fv))
			}
		} else {
			f.SetString(string(fv))
		}
	}
}

// @title Metadata Manifest Index Server
// @version 1.0
// @description Register and search metadata manifest objects
// @contact.name SUNET NOC
// @contact.url https://www.sunet.se/
// @contact.email noc@sunet.se

// @license.name BSD

// StatusResponse example
type StatusResponse struct {
	Size    int    `json:"size" example:"100"`
	Name    string `json:"name" example:"mix"`
	Version string `json:"version" example:"1.0"`
}

// Status godoc
// @Summary Display status and version information
// @Tags status
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /status [get]
func Status(db *datasets.Datasets, w http.ResponseWriter, r *http.Request) {
	sz, _ := db.Bleve.DocCount()
	status := StatusResponse{
		Size:    int(sz),
		Name:    meta.Name(),
		Version: meta.Version(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Register godoc
// @Summary Register a JSON-LD URL with the index server
// @Tags register
// @Accept json
// @Produce json
// @Success 200 {object} manifest.ManifestInfo
// @Router /register [post]
func Register(db *datasets.Datasets, w http.ResponseWriter, r *http.Request) {
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	url := string(b)
	var m manifest.ManifestInfo
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
	db.Bleve.Index(m.ID, m)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// Search godoc
// @Summary Search the index
// @Tags search
// @Param query query string false "query string"
// @Produce json
// @Success 200 {list} []manifest.ManifestInfo
// @Router /search [get]
func Search(db *datasets.Datasets, w http.ResponseWriter, r *http.Request) {
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
		searchResults, err = db.Bleve.Search(search)
		if err != nil {
			log.Fatal(err) // better to give up
		}
	} else {
		sz, _ := db.Bleve.DocCount()
		query := bleve.NewMatchAllQuery()
		search := bleve.NewSearchRequest(query)
		search.Size = int(sz)
		searchResults, err = db.Bleve.Search(search)
		if err != nil {
			log.Fatal(err) // better to give up
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	enc := json.NewEncoder(w)
	comma := false
	w.Write([]byte("["))
	for _, hit := range searchResults.Hits {
		var m manifest.ManifestInfo
		doc, _ := db.Bleve.Document(hit.ID)
		if comma {
			w.Write([]byte(","))
		}
		comma = true
		doc2manifest(doc, &m)
		enc.Encode(m)
	}
	w.Write([]byte("]"))
}

func NewAPI(db *datasets.Datasets) http.Handler {
	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		Status(db, w, r)
	}).Methods("GET")

	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		Register(db, w, r)
	}).Methods("POST")

	router.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		Search(db, w, r)
	}).Methods("GET")

	docs.SwaggerInfo.Host = utils.Getenv("S3_MM_HOST", "localhost:3000")
	docs.SwaggerInfo.BasePath = utils.Getenv("S3_MM_BASEPATH", "/")

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("//"+docs.SwaggerInfo.Host+docs.SwaggerInfo.BasePath+"/swagger/doc.json"),
		httpSwagger.DeepLinking(false),
	))

	return router
}
