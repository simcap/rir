package main

import (
	"flag"
	"log"
	"math"
	"net"
	"sync"

	"github.com/simcap/rir/providers"
	"github.com/simcap/rir/reader"
	"github.com/simcap/rir/rir"
)

func main() {
	country := flag.String("c", "", "2 letters string of the country (ISO 3166)")
	iptype := flag.String("t", "ipv4", "type of IP addresses")
	ipquery := flag.String("q", "", "ip address to which to resolve country")

	providers.CreateCacheDir()

	results := []*rir.Records{}
	for records := range retrieveData() {
		results = append(results, records)
	}

	flag.Parse()

	if query := *ipquery; query != "" {
		for _, region := range results {
			for _, iprecord := range region.Ips {
				if iprecord.Type == "ipv4" {
					ones := 32 - int(math.Log2(float64(iprecord.Value)))
					ipnet := net.IPNet{iprecord.Start, net.CIDRMask(ones, 32)}

					if ipnet.Contains(net.ParseIP(query)) {
						log.Printf("Country %s %s (%d hosts)", iprecord.Cc, iprecord.Start, iprecord.Value)
					}
				}
			}
		}
	} else {

		if *country == "" {
			log.Fatal("Provide a 2 letter code for a country")
		}

		for _, region := range results {
			for _, iprecord := range region.Ips {
				if iprecord.Cc == *country && iprecord.Type == *iptype {
					log.Printf("Country %s %s (%d hosts)\n", *country, iprecord.Start, iprecord.Value)
				}
			}
		}
	}
}

func retrieveData() chan *rir.Records {
	var wg sync.WaitGroup
	ch := make(chan *rir.Records)

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
