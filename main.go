package main

import (
	"flag"
	"log"
	"net"
	"sync"

	"github.com/simcap/rir/providers"
	"github.com/simcap/rir/reader"
	"github.com/simcap/rir/rir"
)

func main() {
	country := flag.String("c", "", "2 letters string of the country (ISO 3166)")
	ipquery := flag.String("q", "", "ip address to which to resolve country")

	providers.CreateCacheDir()

	results := []*rir.Records{}
	for records := range retrieveData() {
		results = append(results, records)
	}

	flag.Parse()
	query := Query{results, *country, *ipquery}

	var all []rir.IpRecord
	if query.IsCountryQuery() {
		all = query.matchOnCountry()
	} else {
		all = query.matchOnIp()
	}
	for _, r := range all {
		log.Print(r)
	}
}

type Query struct {
	data     []*rir.Records
	country  string
	ipstring string
}

func (q *Query) IsCountryQuery() bool {
	return q.country != ""
}

func (q *Query) matchOnCountry() []rir.IpRecord {
	if q.country == "" {
		log.Fatal("Provide a 2 letter code for a country")
	}

	var results []rir.IpRecord
	for _, region := range q.data {
		for _, iprecord := range region.Ips {
			if iprecord.Cc == q.country && iprecord.Type == rir.IPv4 {
				results = append(results, iprecord)
			}
		}
	}
	return results
}

func (q *Query) matchOnIp() []rir.IpRecord {
	if q.ipstring == "" {
		log.Fatal("Provide a ip address")
	}

	var results []rir.IpRecord
	for _, region := range q.data {
		for _, iprecord := range region.Ips {
			if iprecord.Net().Contains(net.ParseIP(q.ipstring)) {
				results = append(results, iprecord)
			}
		}
	}
	return results
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
