package rir

import (
	"bufio"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Summary struct {
	Registry, Type string
	Count          int
}

type Version struct {
	Version                       float64
	Registry, Serial              string
	Records                       int
	StartDate, EndDate, UtcOffset string
}

type Record struct {
	Registry, Cc, Type string
	Value              int
	Date, Status       string
}

type IpRecord struct {
	*Record
	Start net.IP
}

type AsnRecord struct {
	*Record
	Start int
}

type Records struct {
	Version                               float64
	Count, AsnCount, Ipv4Count, Ipv6Count int
	Asns                                  []AsnRecord
	Ips                                   []IpRecord
}

func Parse(r io.Reader) *Records {
	asnRecords := []AsnRecord{}
	ipRecords := []IpRecord{}
	summaries := []Summary{}
	scanner := bufio.NewScanner(r)
	var version *Version

	for scanner.Scan() {
		line := scanner.Text()
		ignoreLine, _ := regexp.Compile("^#|^\\s*$")
		versionLine, _ := regexp.Compile("^\\d+\\.*\\d*")

		if ignoreLine.MatchString(line) {
			continue
		}

		fields := strings.Split(line, "|")

		if versionLine.MatchString(line) {
			version = parseVersionLine(fields)
		} else if strings.HasSuffix(line, "summary") {
			summary := parseSummaryLine(fields)
			summaries = append(summaries, summary)
		} else {
			if strings.HasPrefix(fields[2], "ipv") {
				record := parseIpRecord(fields)
				ipRecords = append(ipRecords, record)
			} else if strings.HasPrefix(fields[2], "asn") {
				record := parseAsnRecord(fields)
				asnRecords = append(asnRecords, record)
			}
		}
	}

	asnCount, ipv4Count, ipv6Count := recordsCountByType(summaries)

	return &(Records{
		Version:   version.Version,
		Count:     version.Records,
		AsnCount:  asnCount,
		Ipv4Count: ipv4Count,
		Ipv6Count: ipv6Count,
		Asns:      asnRecords,
		Ips:       ipRecords,
	})
}

func parseVersionLine(fields []string) *Version {
	version, _ := strconv.ParseFloat(fields[0], 64)
	recordsCount, _ := strconv.Atoi(fields[3])
	return &Version{
		version, fields[1], fields[2], recordsCount,
		fields[4], fields[5], fields[6],
	}
}

func parseSummaryLine(fields []string) Summary {
	count, _ := strconv.Atoi(fields[4])
	return Summary{fields[0], fields[2], count}
}

func recordsCountByType(summaries []Summary) (int, int, int) {
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

func parseIpRecord(fields []string) IpRecord {
	value, _ := strconv.Atoi(fields[4])
	return IpRecord{
		&Record{fields[0], fields[1], fields[2],
			value, fields[5], fields[6]},
		net.ParseIP(fields[3]),
	}
}

func parseAsnRecord(fields []string) AsnRecord {
	value, _ := strconv.Atoi(fields[4])
	asnNumber, _ := strconv.Atoi(fields[3])
	return AsnRecord{
		&Record{fields[0], fields[1], fields[2],
			value, fields[5], fields[6]},
		asnNumber,
	}
}
