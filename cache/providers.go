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
	"regexp"
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

type CachedProvider struct {
	*DefaultProvider
}

func (p *CachedProvider) GetData() io.Reader {
	f := GetDataFile(p.Name())
	defer f.Close()
	finfo, _ := f.Stat()
	if finfo.Size() == 0 || p.isStale(f) {
		data := p.DefaultProvider.GetData()
		CopyDataToFile(data, f)
	}

	f = GetDataFile(p.Name())
	content, _ := ioutil.ReadAll(f)
	return bytes.NewBuffer(content)
}

func (p *CachedProvider) isStale(f *os.File) bool {
	local := p.localMd5(f)
	remote := p.remoteMd5()
	return (local != remote) || (remote == "" && local == "")
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

func (p *CachedProvider) localMd5(f *os.File) string {
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Cannot checksum local file for %s. %s", p.Name(), err)
	}
	sum := md5.Sum(content)
	log.Printf("Local md5 for %s is %x", p.Name(), sum)
	return fmt.Sprintf("%x", sum)
}

func (p *CachedProvider) remoteMd5() string {
	resp, err := http.Get(p.url + ".md5")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != 200 {
		log.Printf("Cannot GET md5 for %s. Call returned %d", p.Name(), status)
		return ""
	}

	md5Response, _ := ioutil.ReadAll(resp.Body)

	matches := regexp.MustCompile("=\\s*(\\w+)\\s*$").FindSubmatch(md5Response)
	if matches == nil {
		log.Print("Cannot regexp match an md5")
		return ""
	}

	log.Printf("Remote md5 for %s is %s", p.Name(), string(matches[1]))
	return string(matches[1])
}

var (
	Providers = []*CachedProvider{
		&CachedProvider{&DefaultProvider{
			"afrinic",
			"http://ftp.apnic.net/stats/afrinic/delegated-afrinic-latest",
		}},
		//&CachedProvider{&DefaultProvider{
		//	"apnic",
		//	"http://ftp.apnic.net/stats/apnic/delegated-apnic-latest",
		//}},
	}
	//		"iana":    "http://ftp.apnic.net/stats/iana/delegated-iana-latest",
	//		"lacnic":  "http://ftp.apnic.net/stats/lacnic/delegated-lacnic-latest",
	//		"ripencc": "http://ftp.apnic.net/stats/ripe-ncc/delegated-ripencc-latest",
)
