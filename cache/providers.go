package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Provider interface {
	Name() string
	Url() string
	GetData() io.Reader
	IsStale() bool
}

type DefaultProvider struct {
	name string
	url  string
}

func (p *DefaultProvider) Name() string {
	return p.name
}

func (p *DefaultProvider) Url() string {
	return p.url
}

func (p *DefaultProvider) IsStale() bool {
	return false
}

func (p *DefaultProvider) GetData() io.Reader {
	log.Printf("Fetching %s data", p.Name())
	response, err := http.Get(p.url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if status := response.StatusCode; status != 200 {
		log.Fatalf("HTTP call returned %d", status)
	}

	content, _ := ioutil.ReadAll(response.Body)

	return bytes.NewBuffer(content)
}

var (
	Providers = []*DefaultProvider{
		&DefaultProvider{
			"afrinic",
			"http://ftp.apnic.net/stats/afrinic/delegated-afrinic-latest",
		},
		&DefaultProvider{
			"apnic",
			"http://ftp.apnic.net/stats/apnic/delegated-apnic-latest",
		},
	}
	//		"iana":    "http://ftp.apnic.net/stats/iana/delegated-iana-latest",
	//		"lacnic":  "http://ftp.apnic.net/stats/lacnic/delegated-lacnic-latest",
	//		"ripencc": "http://ftp.apnic.net/stats/ripe-ncc/delegated-ripencc-latest",
)
