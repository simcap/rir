package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Provider interface {
	Name() string
	GetData() io.Reader
}

type DefaultProvider struct {
	name string
	url  string
}

func (p *DefaultProvider) Name() string {
	return p.name
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

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewBuffer(content)
}

var (
	AllProviders = []*CachedProvider{
		NewCachedProvider(
			"afrinic",
			"http://ftp.apnic.net/stats/afrinic/delegated-afrinic-extended-latest",
		),
		NewCachedProvider(
			"apnic",
			"http://ftp.apnic.net/stats/apnic/delegated-apnic-extended-latest",
		),
		NewCachedProvider(
			"lacnic",
			"http://ftp.apnic.net/stats/lacnic/delegated-lacnic-extended-latest",
		),
		NewCachedProvider(
			"ripencc",
			"http://ftp.apnic.net/stats/ripe-ncc/delegated-ripencc-extended-latest",
		),
		NewCachedProvider(
			"arin",
			"http://ftp.arin.net/pub/stats/arin/delegated-arin-extended-latest",
		),
		//NewCachedProvider(
		//	"iana",
		//	"http://ftp.apnic.net/stats/iana/delegated-iana-latest",
		//),
	}
)
