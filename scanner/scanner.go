package scanner

import (
	"bufio"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/simcap/rir/rir"
)

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

func (cs *ConcurrentScanner) parseVersionLine() rir.Version {
	version, _ := strconv.ParseFloat(cs.fields[0], 64)
	recordsCount, _ := strconv.Atoi(cs.fields[3])
	return rir.Version{
		version, cs.fields[1], cs.fields[2], recordsCount,
		cs.fields[4], cs.fields[5], cs.fields[6],
	}
}

func (cs *ConcurrentScanner) parseSummaryLine() rir.Summary {
	count, _ := strconv.Atoi(cs.fields[4])
	return rir.Summary{cs.fields[0], cs.fields[2], count}
}

func (cs *ConcurrentScanner) parseIpRecord() rir.IpRecord {
	value, _ := strconv.Atoi(cs.fields[4])
	return rir.IpRecord{
		&rir.Record{cs.fields[0], cs.fields[1], cs.fields[2],
			value, cs.fields[5], cs.fields[6]},
		net.ParseIP(cs.fields[3]),
	}
}

func (cs *ConcurrentScanner) parseAsnRecord() rir.AsnRecord {
	value, _ := strconv.Atoi(cs.fields[4])
	asnNumber, _ := strconv.Atoi(cs.fields[3])
	return rir.AsnRecord{
		&rir.Record{cs.fields[0], cs.fields[1], cs.fields[2],
			value, cs.fields[5], cs.fields[6]},
		asnNumber,
	}
}
