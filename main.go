package main

import (
	"log"
	"sync"

	"github.com/simcap/rir/providers"
	"github.com/simcap/rir/reader"
	"github.com/simcap/rir/scanner"
)

func main() {
	providers.CreateCacheDir()

	results := []*scanner.Records{}
	for records := range retrieveData() {
		results = append(results, records)
	}

	for _, result := range results {
		log.Printf("%d %d %d %d", result.Count, result.AsnCount, result.Ipv4Count, result.Ipv6Count)
	}
}

func retrieveData() chan *scanner.Records {
	var wg sync.WaitGroup
	ch := make(chan *scanner.Records)

	for _, provider := range providers.All {
		wg.Add(1)
		go func(p providers.Provider) {
			defer wg.Done()
			records, err := reader.NewReader(p.GetData()).Read()
			if err != nil {
				log.Fatal(err)
			}
			ch <- records
		}(provider)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}
