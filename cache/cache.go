package cache

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func CreateRirCacheDir() {
	for _, provider := range Providers {
		path := filepath.Join(GetRirDir(), provider.Name())
		os.MkdirAll(path, 0700)
	}
}

func Refresh() {
	refresh := make(chan Provider)
	uptodate := make(chan Provider)

	for _, provider := range Providers {
		go func(p Provider) {
			local := localMd5For(p)
			remote := remoteMd5For(p)
			if local != remote || (remote == "" && local == "") {
				refresh <- p
			} else {
				uptodate <- p
			}
		}(provider)
	}

	for range Providers {
		select {
		case r := <-refresh:
			log.Printf("Need to refresh %s", r.Name())
			CopyDataToFile(r.GetData(), GetDataFile(r.Name()))
		case u := <-uptodate:
			log.Printf("%s is up to date", u.Name())
		}
	}
}

func localMd5For(provider Provider) string {
	content, err := ioutil.ReadAll(GetDataFile(provider.Name()))
	if err != nil {
		log.Fatalf("Cannot checksum local file for %s. %s", provider.Name(), err)
	}
	sum := md5.Sum(content)
	log.Printf("Local md5 for %s is %x", provider.Name(), sum)
	return fmt.Sprintf("%x", sum)
}

func remoteMd5For(provider Provider) string {
	resp, err := http.Get(provider.Url() + ".md5")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != 200 {
		log.Printf("Cannot GET md5 for %s. Call returned %d", provider.Name(), status)
		return ""
	}

	md5Response, _ := ioutil.ReadAll(resp.Body)

	matches := regexp.MustCompile("=\\s*(\\w+)\\s*$").FindSubmatch(md5Response)
	if matches == nil {
		log.Print("Cannot regexp match an md5")
		return ""
	}

	log.Printf("Remote md5 for %s is %s", provider.Name(), string(matches[1]))
	return string(matches[1])
}
