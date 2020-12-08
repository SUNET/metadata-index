package utils

import (
	"os"
)

func Getenv(key string, dflt string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return dflt
	} else {
		return v
	}
}
