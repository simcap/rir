package main

import (
	"bufio"
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
	"time"
)

type CachedProvider struct {
	*DefaultProvider
}

func NewCachedProvider(name string, url string) *CachedProvider {
	return &CachedProvider{
		&DefaultProvider{name, url},
	}
}

const ONE_DAY = 86400.0

func (p *CachedProvider) GetData() io.Reader {
	f := p.GetFile()
	defer f.Close()
	finfo, _ := f.Stat()

	fileage := time.Since(finfo.ModTime()).Seconds()

	if finfo.Size() > 0 && fileage < ONE_DAY {
		// no refresh needed
	} else if finfo.Size() == 0 || p.isStale() {
		log.Printf("Refreshing %s data", p.Name())
		data := p.DefaultProvider.GetData()
		CopyDataToFile(data, f)
	}

	content, _ := ioutil.ReadFile(f.Name())
	return bytes.NewBuffer(content)
}

func (p *CachedProvider) isStale() bool {
	local := p.localMd5()
	remote := p.remoteMd5()
	return (local != remote) || (remote == "" && local == "")
}

func GetCacheDir() string {
	return filepath.Join(os.Getenv("HOME"), ".rir")
}

func CreateCacheDir() {
	for _, provider := range AllProviders {
		path := filepath.Join(GetCacheDir(), provider.Name())
		os.MkdirAll(path, 0700)
	}
}

func CopyDataToFile(data io.Reader, dest *os.File) {
	log.Printf("Start copying to %s", dest.Name())
	writer := bufio.NewWriter(dest)
	if _, err := io.Copy(writer, data); err != nil {
		log.Fatal(err)
	}
	err := writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Finished copying to %s", dest.Name())
}

func (p *CachedProvider) GetFile() *os.File {
	f, err := os.OpenFile(p.filePath(), os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func (p *CachedProvider) filePath() string {
	return filepath.Join(GetCacheDir(), p.Name(), "latest")
}

func (p *CachedProvider) localMd5() string {
	content, err := ioutil.ReadFile(p.filePath())
	if err != nil {
		log.Fatalf("Cannot checksum local file for %s. %s", p.Name(), err)
	}
	return fmt.Sprintf("%x", md5.Sum(content))
}

var MD5SigRegex = regexp.MustCompile(`(?i)([a-f0-9]{32})`)

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

	matches := MD5SigRegex.FindSubmatch(md5Response)

	if matches == nil {
		log.Printf("Cannot regexp match an md5 for %s", p.Name())
		return ""
	}

	return string(matches[1])
}
