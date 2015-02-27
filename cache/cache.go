package cache

import (
	"log"
	"os"
	"path/filepath"
)

func CreateRirCacheDir() {
	for _, provider := range Providers {
		path := filepath.Join(GetRirDir(), provider.Name())
		os.MkdirAll(path, 0700)
	}
}

func Refresh() {
	for _, provider := range Providers {
		go func(p Provider) {
			if p.IsStale() {
				log.Printf("Need to refresh %s", p.Name())
				CopyDataToFile(p.GetData(), GetDataFile(p.Name()))
			} else {
				log.Printf("%s is up to date", p.Name())
			}
		}(provider)
	}
}
