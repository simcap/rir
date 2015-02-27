package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/simcap/rir/cache"
	"github.com/simcap/rir/reader"
)

func main() {
	createRirCacheDir()

	collect := make(chan *reader.Records, len(cache.Providers))

	runtime.GOMAXPROCS(runtime.NumCPU())

	for _, provider := range cache.Providers {
		go fetch(provider, collect)
	}

	results := []*reader.Records{}
	for range cache.Providers {
		select {
		case r := <-collect:
			results = append(results, r)
		}
	}

	for _, result := range results {
		log.Printf("%d %d %d %d", result.Count, result.AsnCount, result.Ipv4Count, result.Ipv6Count)
	}
}

func fetch(provider cache.Provider, results chan<- *reader.Records) {
	data := provider.GetData()
	log.Printf("Parsing results for %s", provider.Name())
	records, parseErr := reader.NewReader(data).Read()
	if parseErr != nil {
		log.Fatal(parseErr)
	}
	results <- records
}

func createRirCacheDir() {
	for _, provider := range cache.Providers {
		path := filepath.Join(cache.GetRirDir(), provider.Name())
		os.MkdirAll(path, 0700)
	}
}
