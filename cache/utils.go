package cache

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
)

func GetDataFile(provider string) *os.File {
	path := filepath.Join(GetRirDir(), provider, "latest")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func CopyDataToFile(data io.Reader, dest *os.File) {
	log.Printf("Start copying to %s", dest.Name())
	defer dest.Close()
	writer := bufio.NewWriter(dest)
	if _, err := io.Copy(writer, data); err != nil {
		log.Fatal(err)
	}
	writer.Flush()
}

func GetRirDir() string {
	return filepath.Join(os.Getenv("HOME"), ".rir")
}
