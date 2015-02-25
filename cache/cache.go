package cache

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
)

func CreateRirCacheDir() {
	u, _ := user.Current()
	for provider, _ := range Providers {
		providerDir := filepath.Join(u.HomeDir, ".rir", provider)
		os.MkdirAll(providerDir, 0700)
	}
}

func Refresh() {
	refresh := make(chan string)
	uptodate := make(chan string)

	for provider, _ := range Providers {
		go func(p string) {
			local := LocalMd5For(p)
			remote := RemoteMd5For(p)
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
		case u := <-uptodate:
			log.Printf("%s is up to date", u)
		}
	}
}

func LocalMd5For(provider string) string {
	log.Printf("Verifying local md5 for %s", provider)
	u, _ := user.Current()
	file, err := os.Open(filepath.Join(u.HomeDir, ".rir", provider, "md5"))
	if os.IsNotExist(err) {
		return ""
	}
	content, _ := ioutil.ReadAll(file)
	return string(content)
}

func RemoteMd5For(provider string) string {
	log.Printf("Verifying remote md5 for %s", provider)
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
		return ""
	}

	log.Printf("Remote md5 is '%s'", string(matches[1]))
	return string(matches[1])
}
