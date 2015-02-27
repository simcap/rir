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
	if finfo.Size() == 0 || p.isStale(f) {
		data := p.DefaultProvider.GetData()
		CopyDataToFile(data, f)
	}

	f = p.GetFile()
	content, _ := ioutil.ReadAll(f)
	return bytes.NewBuffer(content)
}

func (p *CachedProvider) isStale(f *os.File) bool {
	local := p.localMd5(f)
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
	writer.Flush()
}

func (p *CachedProvider) GetFile() *os.File {
	path := filepath.Join(GetCacheDir(), p.Name(), "latest")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		log.Fatal(err)
	}
	return f
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
