package reader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type ParseError struct {
	Line int
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Err)
}

var (
	ErrMultipleVersionLine = errors.New("redundant version line found")
)

type Reader struct {
	scanner       *bufio.Scanner
	lineCount     int
	currentLine   string
	fields        []string
	versionParsed bool
}

func NewReader(r io.Reader) *Reader {
	return &Reader{scanner: bufio.NewScanner(r)}
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
	asnRecords := []AsnRecord{}
	ipRecords := []IpRecord{}
	summaries := []Summary{}
	var version *Version

	for r.scanner.Scan() {
		line := r.scanner.Text()

		r.currentLine = line
		r.fields = strings.Split(line, "|")
		r.lineCount++

		if r.ignoredLine() {
			continue
		}

		if r.versionLine() {
			if r.versionParsed {
				return nil, r.error(ErrMultipleVersionLine)
			}
			version = r.parseVersionLine()
		} else if r.summaryLine() {
			summary := r.parseSummaryLine()
			summaries = append(summaries, summary)
		} else {
			if r.ipvLine() {
				record := r.parseIpRecord()
				ipRecords = append(ipRecords, record)
			} else if r.asnLine() {
				record := r.parseAsnRecord()
				asnRecords = append(asnRecords, record)
			}
		}
	}

	asnCount, ipv4Count, ipv6Count := r.recordsCountByType(summaries)

	return &(Records{
		Version:   version.Version,
		Count:     version.Records,
		AsnCount:  asnCount,
		Ipv4Count: ipv4Count,
		Ipv6Count: ipv6Count,
		Asns:      asnRecords,
		Ips:       ipRecords,
	}), nil
}

func (r *Reader) error(err error) error {
	return &ParseError{Line: r.lineCount, Err: err}
}

func (r *Reader) versionLine() bool {
	version := regexp.MustCompile("^\\d+\\.*\\d*")
	return version.MatchString(r.currentLine)
}

func (r *Reader) ignoredLine() bool {
	ignored := regexp.MustCompile("^#|^\\s*$")
	return ignored.MatchString(r.currentLine)
}

func (r *Reader) summaryLine() bool {
	return strings.HasSuffix(r.currentLine, "summary")
}

func (r *Reader) ipvLine() bool {
	return strings.HasPrefix(r.fields[2], "ipv")
}

func (r *Reader) asnLine() bool {
	return strings.HasPrefix(r.fields[2], "asn")
}

func (r *Reader) parseVersionLine() *Version {
	version, _ := strconv.ParseFloat(r.fields[0], 64)
	recordsCount, _ := strconv.Atoi(r.fields[3])
	r.versionParsed = true
	return &Version{
		version, r.fields[1], r.fields[2], recordsCount,
		r.fields[4], r.fields[5], r.fields[6],
	}
}

func (r *Reader) parseSummaryLine() Summary {
	count, _ := strconv.Atoi(r.fields[4])
	return Summary{r.fields[0], r.fields[2], count}
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

func (r *Reader) parseIpRecord() IpRecord {
	value, _ := strconv.Atoi(r.fields[4])
	return IpRecord{
		&Record{r.fields[0], r.fields[1], r.fields[2],
			value, r.fields[5], r.fields[6]},
		net.ParseIP(r.fields[3]),
	}
}

func (r *Reader) parseAsnRecord() AsnRecord {
	value, _ := strconv.Atoi(r.fields[4])
	asnNumber, _ := strconv.Atoi(r.fields[3])
	return AsnRecord{
		&Record{r.fields[0], r.fields[1], r.fields[2],
			value, r.fields[5], r.fields[6]},
		asnNumber,
	}
}
