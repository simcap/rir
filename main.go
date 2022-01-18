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
	hostscount := flag.Bool("n", false, "given country return possible hosts count (exclude network and broadcast addresses)")

	CreateCacheDir()

	var data []*Records
	for records := range retrieveData() {
		data = append(data, records)
	}

	flag.Parse()
	query := Query{data, *country, *ipquery, *hostscount}

	if query.IsCountryQuery() {
		results := query.matchOnCountry()
		if query.hostscount {
			var count int
			for _, r := range results {
				ones, size := r.Mask.Size()
				mask := size - ones
				if mask > 0 {
					count = count + int((math.Pow(2, float64(mask)) - 2))
				}
			}
			fmt.Println(count)
		} else {
			for _, r := range results {
				fmt.Println(r)
			}
		}
	} else {
		for _, r := range query.matchOnIp() {
			fmt.Println(r)
		}
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

func (q *Query) matchOnCountry() []*net.IPNet {
	if q.country == "" {
		flag.Usage()
	}

	var results []*net.IPNet
	for _, region := range q.data {
		for _, iprecord := range region.Ips {
			if iprecord.Cc == q.country && iprecord.Type == IPv4 {
				results = append(results, iprecord.Net()...)
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
			for _, ipnet := range iprecord.Net() {
				if ipnet.Contains(net.ParseIP(q.ipstring)) {
					results = append(results, fmt.Sprintf("%s %s", iprecord.Cc, ipnet))
				}
			}
		}
	}
	return results
}

func retrieveData() chan *Records {
	var wg sync.WaitGroup
	ch := make(chan *Records)
	wg.Add(len(AllProviders))

	for _, provider := range AllProviders {
		go func(p Provider) {
			records, err := NewReader(p.GetData()).Read()
			if err != nil {
				log.Fatal(err)
			}
			ch <- records
			wg.Done()
		}(provider)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}
