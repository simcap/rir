package main

import (
	"log"
	"runtime"

	"github.com/simcap/rir/providers"
	"github.com/simcap/rir/reader"
)

func main() {
	log.Printf("Numbers of CPU %d", runtime.NumCPU())
	//runtime.GOMAXPROCS(runtime.NumCPU())
	providers.CreateCacheDir()

	collect := make(chan *reader.Records, len(providers.All))

	runtime.GOMAXPROCS(runtime.NumCPU())

	for _, provider := range providers.All {
		go fetch(provider, collect)
	}

	results := []*reader.Records{}
	for range providers.All {
		select {
		case r := <-collect:
			results = append(results, r)
		}
	}

	for _, result := range results {
		log.Printf("%d %d %d %d", result.Count, result.AsnCount, result.Ipv4Count, result.Ipv6Count)
	}
}

func fetch(provider providers.Provider, results chan<- *reader.Records) {
	log.Printf("Parsing %s data", provider.Name())
	data := provider.GetData()
	records, parseErr := reader.NewReader(data).Read()
	if parseErr != nil {
		log.Fatal(parseErr)
	}
	results <- records
}
