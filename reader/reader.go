package reader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrMultipleVersionLine = errors.New("multiple version line found")
)

type Reader struct {
	scanners []*ConcurrentScanner
}

func NewReader(r io.Reader) *Reader {
	scans := []*ConcurrentScanner{}
	for _, r := range splitReader(r, 5) {
		scans = append(scans, NewConcurrentScanner(r))
	}
	return &Reader{scanners: scans}
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

type (
	Summary struct {
		Registry, Type string
		Count          int
	}

	Version struct {
		Version                       float64
		Registry, Serial              string
		Records                       int
		StartDate, EndDate, UtcOffset string
	}

	Record struct {
		Registry, Cc, Type string
		Value              int
		Date, Status       string
	}

	IpRecord struct {
		*Record
		Start net.IP
	}

	AsnRecord struct {
		*Record
		Start int
	}

	Records struct {
		Version                               float64
		Count, AsnCount, Ipv4Count, Ipv6Count int
		Asns                                  []AsnRecord
		Ips                                   []IpRecord
	}
)

func (r *Reader) Read() (records *Records, err error) {
	out := make(chan interface{})
	for _, scanner := range r.scanners {
		go func(s *ConcurrentScanner) {
			q := s.Run()
			for n := range q {
				out <- n
			}
			out <- true
		}(scanner)
	}

	asnRecords := []AsnRecord{}
	ipRecords := []IpRecord{}
	summaries := []Summary{}
	var version Version

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
			case Summary:
				summaries = append(summaries, v)
			case IpRecord:
				ipRecords = append(ipRecords, v)
			case AsnRecord:
				asnRecords = append(asnRecords, v)
			case Version:
				version = v
			default:
				log.Fatalf("Do not know this type %T", v)
			}

		}
	}
	close(out)

	asnCount, ipv4Count, ipv6Count := r.recordsCountByType(summaries)

	return &Records{
		Version:   version.Version,
		Count:     version.Records,
		AsnCount:  asnCount,
		Ipv4Count: ipv4Count,
		Ipv6Count: ipv6Count,
		Asns:      asnRecords,
		Ips:       ipRecords,
	}, nil
}

type ConcurrentScanner struct {
	scanner     *bufio.Scanner
	currentLine string
	fields      []string
}

func NewConcurrentScanner(r io.Reader) *ConcurrentScanner {
	return &ConcurrentScanner{scanner: bufio.NewScanner(r)}
}

func (cs *ConcurrentScanner) Run() chan interface{} {
	c := make(chan interface{})

	go func() {
		for cs.scanner.Scan() {
			cs.currentLine = cs.scanner.Text()
			cs.fields = strings.Split(cs.currentLine, "|")

			if cs.ignoredLine() {
				continue
			}

			if cs.versionLine() {
				c <- cs.parseVersionLine()
			} else if cs.summaryLine() {
				c <- cs.parseSummaryLine()
			} else {
				if cs.ipvLine() {
					c <- cs.parseIpRecord()
				} else if cs.asnLine() {
					c <- cs.parseAsnRecord()
				}
			}
		}
		close(c)
	}()
	return c

}

func (cs *ConcurrentScanner) versionLine() bool {
	version := regexp.MustCompile("^\\d+\\.*\\d*")
	return version.MatchString(cs.currentLine)
}

func (cs *ConcurrentScanner) ignoredLine() bool {
	ignored := regexp.MustCompile("^#|^\\s*$")
	return ignored.MatchString(cs.currentLine)
}

func (cs *ConcurrentScanner) summaryLine() bool {
	return strings.HasSuffix(cs.currentLine, "summary")
}

func (cs *ConcurrentScanner) ipvLine() bool {
	return strings.HasPrefix(cs.fields[2], "ipv")
}

func (cs *ConcurrentScanner) asnLine() bool {
	return strings.HasPrefix(cs.fields[2], "asn")
}

func (cs *ConcurrentScanner) parseVersionLine() Version {
	version, _ := strconv.ParseFloat(cs.fields[0], 64)
	recordsCount, _ := strconv.Atoi(cs.fields[3])
	return Version{
		version, cs.fields[1], cs.fields[2], recordsCount,
		cs.fields[4], cs.fields[5], cs.fields[6],
	}
}

func (cs *ConcurrentScanner) parseSummaryLine() Summary {
	count, _ := strconv.Atoi(cs.fields[4])
	return Summary{cs.fields[0], cs.fields[2], count}
}

func (r *Reader) recordsCountByType(summaries []Summary) (int, int, int) {
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

func (cs *ConcurrentScanner) parseIpRecord() IpRecord {
	value, _ := strconv.Atoi(cs.fields[4])
	return IpRecord{
		&Record{cs.fields[0], cs.fields[1], cs.fields[2],
			value, cs.fields[5], cs.fields[6]},
		net.ParseIP(cs.fields[3]),
	}
}

func (cs *ConcurrentScanner) parseAsnRecord() AsnRecord {
	value, _ := strconv.Atoi(cs.fields[4])
	asnNumber, _ := strconv.Atoi(cs.fields[3])
	return AsnRecord{
		&Record{cs.fields[0], cs.fields[1], cs.fields[2],
			value, cs.fields[5], cs.fields[6]},
		asnNumber,
	}
}
