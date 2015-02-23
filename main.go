package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/simcap/rir/reader"
)

var (
	logger = log.New(os.Stdout, "", log.Ltime)

	providers = map[string]string{
		"afrinic": "http://ftp.apnic.net/stats/afrinic/delegated-afrinic-latest",
		"apnic":   "http://ftp.apnic.net/stats/apnic/delegated-apnic-latest",
		"iana":    "http://ftp.apnic.net/stats/iana/delegated-iana-latest",
		"lacnic":  "http://ftp.apnic.net/stats/lacnic/delegated-lacnic-latest",
		"ripencc": "http://ftp.apnic.net/stats/ripe-ncc/delegated-ripencc-latest",
	}
)

func main() {
	createRirHomeDir()

	refresh := make(chan string)
	uptodate := make(chan string)

	for provider, _ := range providers {
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

	for i := 0; i < len(providers); i++ {
		select {
		case r := <-refresh:
			logger.Printf("Need to refresh %s", r)
		case u := <-uptodate:
			logger.Printf("%s is up to date", u)
		}
	}

	collect := make(chan *reader.Records, len(providers))

	runtime.GOMAXPROCS(runtime.NumCPU())

	for provider, url := range providers {
		go fetch(provider, url, collect)
	}

	results := []*reader.Records{<-collect, <-collect, <-collect, <-collect}

	for _, result := range results {
		logger.Printf("%d %d %d %d", result.Count, result.AsnCount, result.Ipv4Count, result.Ipv6Count)
	}
}

func createRirHomeDir() {
	u, _ := user.Current()
	for provider, _ := range providers {
		providerDir := filepath.Join(u.HomeDir, ".rir", provider)
		os.MkdirAll(providerDir, 0700)
	}
}

func localMd5For(provider string) string {
	logger.Printf("Verifying local md5 for %s", provider)
	u, _ := user.Current()
	file, err := os.Open(filepath.Join(u.HomeDir, ".rir", provider, "md5"))
	if os.IsNotExist(err) {
		return ""
	}
	content, _ := ioutil.ReadAll(file)
	return string(content)
}

func remoteMd5For(provider string) string {
	logger.Printf("Verifying remote md5 for %s", provider)
	resp, err := http.Get(providers[provider] + ".md5")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != 200 {
		logger.Printf("Cannot GET md5 for %s. Call returned %d", provider, status)
		return ""
	}

	md5Response, _ := ioutil.ReadAll(resp.Body)

	matches := regexp.MustCompile("=\\s*(\\w+)\\s*$").FindSubmatch(md5Response)
	if matches == nil {
		return ""
	}

	logger.Printf("Remote md5 is '%s'", string(matches[1]))
	return string(matches[1])
}

func fetch(provider string, url string, results chan<- *reader.Records) {
	logger.Printf("Fetching %s data", provider)
	response, err := http.Get(url)
	if err != nil {
		logger.Fatal(err)
	}
	defer response.Body.Close()

	if status := response.StatusCode; status != 200 {
		logger.Fatalf("GET call returned %d", status)
	}

	logger.Printf("Parsing results for %s", provider)
	records, parseErr := reader.NewReader(response.Body).Read()
	if parseErr != nil {
		logger.Fatal(parseErr)
	}
	results <- records
}
