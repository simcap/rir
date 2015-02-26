package cache

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func CreateRirCacheDir() {
	for provider, _ := range Providers {
		path := filepath.Join(GetRirDir(), provider)
		os.MkdirAll(path, 0700)
	}
}

func Refresh() {
	refresh := make(chan string)
	uptodate := make(chan string)

	for provider, _ := range Providers {
		go func(p string) {
			local := localMd5For(p)
			remote := remoteMd5For(p)
			if local != remote || (remote == "" && local == "") {
				refresh <- p
			} else {
				uptodate <- p
			}
		}(provider)
	}

	for i := 0; i < len(Providers); i++ {
		select {
		case r := <-refresh:
			log.Printf("Need to refresh %s", r)
			CopyDataToFile(Fetch(r), GetDataFile(r))
		case u := <-uptodate:
			log.Printf("%s is up to date", u)
		}
	}
}

func Fetch(provider string) io.Reader {
	log.Printf("Fetching %s data", provider)
	response, err := http.Get(Providers[provider])
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

func localMd5For(provider string) string {
	content, err := ioutil.ReadAll(GetDataFile(provider))
	if err != nil {
		log.Fatalf("Cannot checksum local file for %s. %s", provider, err)
	}
	sum := md5.Sum(content)
	log.Printf("Local md5 for %s is %x", provider, sum)
	return fmt.Sprintf("%x", sum)
}

func remoteMd5For(provider string) string {
	resp, err := http.Get(Providers[provider] + ".md5")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != 200 {
		log.Printf("Cannot GET md5 for %s. Call returned %d", provider, status)
		return ""
	}

	md5Response, _ := ioutil.ReadAll(resp.Body)

	matches := regexp.MustCompile("=\\s*(\\w+)\\s*$").FindSubmatch(md5Response)
	if matches == nil {
		log.Print("Cannot regexp match an md5")
		return ""
	}

	log.Printf("Remote md5 for %s is %s", provider, string(matches[1]))
	return string(matches[1])
}
