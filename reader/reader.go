package reader

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"

	"github.com/simcap/rir/scanner"
)

type Reader struct {
	scanners []*scanner.ConcurrentScanner
}

func NewReader(r io.Reader) *Reader {
	scans := []*scanner.ConcurrentScanner{}
	for _, r := range splitReader(r, 5) {
		scans = append(scans, scanner.NewConcurrentScanner(r))
	}
	return &Reader{scanners: scans}
}

func (r *Reader) Read() (records *scanner.Records, err error) {
	out := make(chan interface{})
	for _, scan := range r.scanners {
		go func(s *scanner.ConcurrentScanner) {
			q := s.Run()
			for n := range q {
				out <- n
			}
			out <- true
		}(scan)
	}

	asnRecords := []scanner.AsnRecord{}
	ipRecords := []scanner.IpRecord{}
	summaries := []scanner.Summary{}
	var version scanner.Version

	dones := []bool{}
Loop:
	for {
		select {
		case value := <-out:
			switch v := value.(type) {
			case bool:
				dones = append(dones, v)
				if len(dones) == len(r.scanners) {
					break Loop
				}
			case scanner.Summary:
				summaries = append(summaries, v)
			case scanner.IpRecord:
				ipRecords = append(ipRecords, v)
			case scanner.AsnRecord:
				asnRecords = append(asnRecords, v)
			case scanner.Version:
				version = v
			default:
				log.Fatalf("Do not know this type %T", v)
			}

		}
	}
	close(out)

	asnCount, ipv4Count, ipv6Count := r.recordsCountByType(summaries)

	return &scanner.Records{
		Version:   version.Version,
		Count:     version.Records,
		AsnCount:  asnCount,
		Ipv4Count: ipv4Count,
		Ipv6Count: ipv6Count,
		Asns:      asnRecords,
		Ips:       ipRecords,
	}, nil
}
func (r *Reader) recordsCountByType(summaries []scanner.Summary) (int, int, int) {
	var asn, ipv4, ipv6 int
	for _, current := range summaries {
		if current.Type == "asn" {
			asn = current.Count
		}
		if current.Type == "ipv4" {
			ipv4 = current.Count
		}
		if current.Type == "ipv6" {
			ipv6 = current.Count
		}
	}
	return asn, ipv4, ipv6
}

func splitReader(r io.Reader, number int) []io.Reader {
	var content []byte
	if v, ok := r.(*bytes.Buffer); ok {
		content = v.Bytes()
	} else {
		content, _ = ioutil.ReadAll(r)
	}

	totalSize := len(content)
	sectionSize := totalSize / number

	readers := []io.Reader{}
	var start, toread, index int
	for i := 1; i <= number; i++ {
		if i < number {
			index = nextNewlineIndex(content, i*sectionSize)
			toread = index - start
		} else {
			toread = totalSize - start
		}

		readers = append(readers,
			io.NewSectionReader(bytes.NewReader(content), int64(start), int64(toread)),
		)
		start = index
	}

	return readers
}

const searchFor int = 100

func nextNewlineIndex(content []byte, start int) int {
	buff := bytes.NewBuffer(content[start:len(content)])
	index := -1
	for i := 1; i < searchFor; i++ {
		b := buff.Next(1)
		if b[0] == '\n' {
			index = i
		}

	}
	if index == -1 {
		log.Fatalf("Could not find newline in next %d bytes", searchFor)
		return index
	} else {
		return index + start
	}
}
