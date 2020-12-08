package main

import (
	//"fmt"
        "github.com/blevesearch/bleve"
        "github.com/blevesearch/bleve/analysis/analyzer/custom"
        "github.com/blevesearch/bleve/analysis/analyzer/keyword"
        "github.com/blevesearch/bleve/analysis/token/edgengram"
        "github.com/blevesearch/bleve/analysis/token/lowercase"
        "github.com/blevesearch/bleve/analysis/token/stop"
        "github.com/blevesearch/bleve/analysis/tokenizer/unicode"
        "github.com/blevesearch/bleve/analysis/tokenmap"
        blevemapping "github.com/blevesearch/bleve/mapping"
        "github.com/sirupsen/logrus"
	"sync"
        //"io"
        //"os"
        //"strings"
        //"time"
)

var log = logrus.New()

type Datasets struct {
	index		bleve.Index
	indexPath	string
}

func NewDatasets(indexPath string) Datasets {
	ds := &Datasets {indexPath: indexPath }
	return *ds
}

type TypedSchema struct {
	ID		string `json:"@id"`
	Schema		string `json:"mm:schema"`
}

type ManifestInfo struct {
	ID		string `json:"@id"`
	Publisher	string `json:"mm:publisher"`
	Creator		string `json:"mm:creator"`
	RightsHolder	string `json:"mm:rightsHolder"`
        Manifest	[]TypedSchema `json:"mm:manifest"`
}

func NgramTokenFilter() map[string]interface{} {
        return map[string]interface{}{
                "type": edgengram.Name,
                "back": false,
                "min":  3.0,
                "max":  25.0,
        }
}

func StopWordsTokenMap() map[string]interface{} {
        return map[string]interface{}{
                "type": tokenmap.Name,
                "tokens": []interface{}{
                        "a", "an", "of", "the", "die", "von", "av", "i", "identity", "provider", "university", "uni",
                },
        }
}

func StopWordsTokenFilter() map[string]interface{} {
        return map[string]interface{}{
                "type":           stop.Name,
                "stop_token_map": "stop_words_map",
        }
}

func NgramAnalyzer() map[string]interface{} {
        return map[string]interface{}{
                "type":      custom.Name,
                "tokenizer": unicode.Name,
                "token_filters": []string{
                        lowercase.Name,
                        "stop_words_filter",
                        "ngram_tokenfilter",
                },
        }
}

func NewIndexMapping() blevemapping.IndexMapping {
        var err error
        mapping := bleve.NewIndexMapping()
        mapping.DefaultType = "entity"
        err = mapping.AddCustomTokenMap("stop_words_map", StopWordsTokenMap())
        if err != nil {
                log.Fatal(err)
        }
        err = mapping.AddCustomTokenFilter("stop_words_filter", StopWordsTokenFilter())
        if err != nil {
                log.Fatal(err)
        }
        err = mapping.AddCustomTokenFilter("ngram_tokenfilter", NgramTokenFilter())
        if err != nil {
                log.Fatal(err)
        }
        err = mapping.AddCustomAnalyzer("ngram_analyzer", NgramAnalyzer())
        if err != nil {
                log.Fatal(err)
        }
        manifestMapping := bleve.NewDocumentMapping()
        mapping.AddDocumentMapping("manifest", manifestMapping)

        nopFieldMapping := bleve.NewTextFieldMapping()
        nopFieldMapping.Analyzer = keyword.Name
        nopFieldMapping.IncludeTermVectors = false
        nopFieldMapping.IncludeInAll = false
        nopFieldMapping.DocValues = false
        nopFieldMapping.Store = false
        nopFieldMapping.Index = false

        contentFieldMapping := bleve.NewTextFieldMapping()
        contentFieldMapping.Analyzer = "ngram_analyzer"
        manifestMapping.AddFieldMappingsAt("Publisher", contentFieldMapping)
	manifestMapping.AddFieldMappingsAt("Creator", contentFieldMapping)
	manifestMapping.AddFieldMappingsAt("RightsHolder", contentFieldMapping)

        manifestMapping.AddFieldMappingsAt("ID", nopFieldMapping)
        return mapping
}

func (db *Datasets) LoadIndex(wg *sync.WaitGroup) {
	var err error
	if db.index == nil {
                db.index, err = bleve.Open(db.indexPath)
                if err == bleve.ErrorIndexPathDoesNotExist {
                        db.index, err = bleve.New(db.indexPath, NewIndexMapping())
                        if err != nil {
                                log.Fatalf("failed to create index: %s", err)
                        }
                } else if err != nil {
                        log.Fatal(err)
                }
        }
}

func (db *Datasets) Reload() {
        var wg sync.WaitGroup
        db.LoadIndex(&wg)
        wg.Wait()
}
