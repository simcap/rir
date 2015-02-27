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
	refresh := make(chan Provider)
	uptodate := make(chan Provider)

	for _, provider := range Providers {
		go func(p Provider) {
			if p.IsStale() {
				refresh <- p
			} else {
				uptodate <- p
			}
		}(provider)
	}

	for range Providers {
		select {
		case r := <-refresh:
			log.Printf("Need to refresh %s", r.Name())
			CopyDataToFile(r.GetData(), GetDataFile(r.Name()))
		case u := <-uptodate:
			log.Printf("%s is up to date", u.Name())
		}
	}
}
