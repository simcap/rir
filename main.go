package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/simcap/rir/cache"
	"github.com/simcap/rir/reader"
)

func main() {
	cache.CreateRirCacheDir()
	cache.Refresh()

	collect := make(chan *reader.Records, len(cache.Providers))

	runtime.GOMAXPROCS(runtime.NumCPU())

	for provider, url := range cache.Providers {
		go fetch(provider, url, collect)
	}

	results := []*reader.Records{<-collect, <-collect, <-collect, <-collect}

	for _, result := range results {
		log.Printf("%d %d %d %d", result.Count, result.AsnCount, result.Ipv4Count, result.Ipv6Count)
	}
}

func fetch(provider string, url string, results chan<- *reader.Records) {
	log.Printf("Fetching %s data", provider)
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if status := response.StatusCode; status != 200 {
		log.Fatalf("GET call returned %d", status)
	}

	log.Printf("Parsing results for %s", provider)
	records, parseErr := reader.NewReader(response.Body).Read()
	if parseErr != nil {
		log.Fatal(parseErr)
	}
	results <- records
}
