package providers

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

func (p *CachedProvider) GetData() io.Reader {
	f := p.GetFile()
	defer f.Close()
	finfo, _ := f.Stat()

	fileage := time.Since(finfo.ModTime()).Seconds()

	if finfo.Size() > 0 && fileage < 86400.0 {
		log.Printf("No refresh for %s", p.Name())
	} else if finfo.Size() == 0 || p.isStale() {
		log.Printf("%s data need refresh", p.Name())
		data := p.DefaultProvider.GetData()
		CopyDataToFile(data, f)
	}

	content, _ := ioutil.ReadFile(f.Name())
	return bytes.NewBuffer(content)
}

func (p *CachedProvider) isStale() bool {
	log.Printf("Checking freshness for %s", p.Name())
	local := p.localMd5()
	remote := p.remoteMd5()
	return (local != remote) || (remote == "" && local == "")
}

func GetCacheDir() string {
	return filepath.Join(os.Getenv("HOME"), ".rir")
}

func CreateCacheDir() {
	for _, provider := range All {
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

	return string(matches[1])
}
