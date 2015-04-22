package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
)

func main() {
	country := flag.String("c", "", "2 letters string of the country (ISO 3166)")
	ipquery := flag.String("q", "", "ip address to which to resolve country")

	CreateCacheDir()

	var data []*Records
	for records := range retrieveData() {
		data = append(data, records)
	}

	flag.Parse()
	query := Query{data, *country, *ipquery}

	var results []string
	if query.IsCountryQuery() {
		results = query.matchOnCountry()
	} else {
		results = query.matchOnIp()
	}
	for _, r := range results {
		fmt.Println(r)
	}
}

type Query struct {
	data     []*Records
	country  string
	ipstring string
}

func (q *Query) IsCountryQuery() bool {
	return q.country != ""
}

func (q *Query) matchOnCountry() []string {
	if q.country == "" {
		flag.Usage()
	}

	var results []string
	for _, region := range q.data {
		for _, iprecord := range region.Ips {
			if iprecord.Cc == q.country && iprecord.Type == IPv4 {
				results = append(results, iprecord.Net().String())
			}
		}
	}
	return results
}

func (q *Query) matchOnIp() []string {
	if q.ipstring == "" {
		flag.Usage()
	}

	var results []string
	for _, region := range q.data {
		for _, iprecord := range region.Ips {
			ipnet := iprecord.Net()
			if ipnet.Contains(net.ParseIP(q.ipstring)) {
				results = append(results, fmt.Sprintf("%s %s", iprecord.Cc, ipnet))
			}
		}
	}
	return results
}

func retrieveData() chan *Records {
	var wg sync.WaitGroup
	ch := make(chan *Records)

	for _, provider := range AllProviders {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			records, err := NewReader(p.GetData()).Read()
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
