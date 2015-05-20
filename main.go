package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
)

func main() {
	country := flag.String("c", "", "2 letters string of the country (ISO 3166)")
	ipquery := flag.String("q", "", "ip address to which to resolve country")
	hostscount := flag.Bool("n", false, "count the number of possibles hosts for a given country")

	CreateCacheDir()

	var data []*Records
	for records := range retrieveData() {
		data = append(data, records)
	}

	flag.Parse()
	query := Query{data, *country, *ipquery, *hostscount}

	var results []interface{}
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
	data       []*Records
	country    string
	ipstring   string
	hostscount bool
}

func (q *Query) IsCountryQuery() bool {
	return q.country != ""
}

func (q *Query) matchOnCountry() []interface{} {
	if q.country == "" {
		flag.Usage()
	}

	var results []interface{}
	for _, region := range q.data {
		for _, iprecord := range region.Ips {
			if iprecord.Cc == q.country && iprecord.Type == IPv4 {
				results = append(results, iprecord.Net())
			}
		}
	}
	if q.hostscount {
		var count int
		for _, r := range results {
			n := r.(*net.IPNet)
			ones, size := n.Mask.Size()
			mask := size - ones
			if mask > 0 {
				count = count + int((math.Pow(2, float64(mask)) - 2))
			}

		}
		return []interface{}{count}
	}
	return results
}

func (q *Query) matchOnIp() []interface{} {
	if q.ipstring == "" {
		flag.Usage()
	}

	var results []interface{}
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
