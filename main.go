package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/simcap/rir/reader"
)

var logger = log.New(os.Stdout, "", log.Ltime)

func main() {
	collect := make(chan *reader.Records, 4)

	go fetch("afrinic", collect)
	go fetch("apnic", collect)
	go fetch("iana", collect)
	go fetch("lacnic", collect)

	results := []*reader.Records{<-collect, <-collect, <-collect, <-collect}

	for _, result := range results {
		logger.Printf("%d %d %d %d", result.Count, result.AsnCount, result.Ipv4Count, result.Ipv6Count)
	}
}

func fetch(provider string, results chan<- *reader.Records) {
	url := fmt.Sprintf("http://ftp.apnic.net/stats/%s/delegated-%s-latest", provider, provider)

	logger.Printf("Fetching %s data", provider)
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if status := response.StatusCode; status != 200 {
		logger.Fatalf("GET call returned %d", status)
	}

	logger.Printf("Parsing %s results", provider)
	records, parseErr := reader.NewReader(response.Body).Read()
	if parseErr != nil {
		log.Fatal(parseErr)
	}
	results <- records
}
