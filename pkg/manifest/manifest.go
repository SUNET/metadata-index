package manifest

import (
        "github.com/sirupsen/logrus"
)

var log = logrus.New()

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
