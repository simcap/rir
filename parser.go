package rir

import (
	"bufio"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type SummaryLine struct {
	Registry, Type string
	Count          int
}

type VersionLine struct {
	Version                       int
	Registry, Serial              string
	Records                       int
	StartDate, EndDate, UtcOffset string
}

type RecordLine struct {
	Registry, Cc, Type string
	Value              int
	Date, Status       string
}

type IpRecord struct {
	*RecordLine
	Start net.IP
}

type AsnRecord struct {
	*RecordLine
	Start int
}

type Records struct {
	Count, AsnCount, Ipv4Count, Ipv6Count int
	Asns                                  []AsnRecord
	Ips                                   []IpRecord
}

func Parse(r io.Reader) *Records {
	asnRecords := []AsnRecord{}
	ipRecords := []IpRecord{}
	summaries := []SummaryLine{}
	scanner := bufio.NewScanner(r)
	var versionLine *VersionLine

	for scanner.Scan() {
		line := scanner.Text()
		ignoreLine, _ := regexp.Compile("^#|^\\s*$")

		if ignoreLine.MatchString(line) {
			continue
		}

		fields := strings.Split(line, "|")

		if first, _ := utf8.DecodeRuneInString(fields[0]); unicode.IsDigit(first) {
			versionLine = parseVersionLine(fields)
		} else if strings.HasSuffix(line, "summary") {
			summary := parseSummaryLine(fields)
			summaries = append(summaries, summary)
		} else {
			if strings.HasPrefix(fields[2], "ipv") {
				record := parseIpRecordLine(fields)
				ipRecords = append(ipRecords, record)
			} else if strings.HasPrefix(fields[2], "asn") {
				record := parseAsnRecordLine(fields)
				asnRecords = append(asnRecords, record)
			}
		}
	}

	var asnCount, ipv4Count, ipv6Count int
	for _, current := range summaries {
		if current.Type == "asn" {
			asnCount = current.Count
		}
		if current.Type == "ipv4" {
			ipv4Count = current.Count
		}
		if current.Type == "ipv6" {
			ipv6Count = current.Count
		}
	}

	return &(Records{
		Count:     versionLine.Records,
		AsnCount:  asnCount,
		Ipv4Count: ipv4Count,
		Ipv6Count: ipv6Count,
		Asns:      asnRecords,
		Ips:       ipRecords,
	})
}

func parseVersionLine(fields []string) *VersionLine {
	version, _ := strconv.Atoi(fields[0])
	recordsCount, _ := strconv.Atoi(fields[3])
	return &VersionLine{
		version, fields[1], fields[2], recordsCount,
		fields[4], fields[5], fields[6],
	}
}

func parseSummaryLine(fields []string) SummaryLine {
	count, _ := strconv.Atoi(fields[4])
	return SummaryLine{fields[0], fields[2], count}
}

func parseIpRecordLine(fields []string) IpRecord {
	count, _ := strconv.Atoi(fields[4])
	return IpRecord{
		&RecordLine{fields[0], fields[1], fields[2],
			count, fields[5], fields[6]},
		net.ParseIP(fields[3]),
	}
}

func parseAsnRecordLine(fields []string) AsnRecord {
	count, _ := strconv.Atoi(fields[4])
	asnNumber, _ := strconv.Atoi(fields[3])
	return AsnRecord{
		&RecordLine{fields[0], fields[1], fields[2],
			count, fields[5], fields[6]},
		asnNumber,
	}
}
