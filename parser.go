package rir

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Summary struct {
	Registry, Type string
	Count          int
}

type VersionLine struct {
	Version                       int
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
	Count, AsnCount, Ipv4Count, Ipv6Count int
	Asns                                  []AsnRecord
	Ips                                   []IpRecord
}

func Parse(r io.Reader) *Records {
	asnRecords := []AsnRecord{}
	ipRecords := []IpRecord{}
	summaries := []Summary{}
	scanner := bufio.NewScanner(r)
	var versionLine VersionLine

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "|")

		if first, _ := utf8.DecodeRuneInString(fields[0]); unicode.IsDigit(first) {
			version, _ := strconv.Atoi(fields[0])
			recordsCount, _ := strconv.Atoi(fields[3])
			versionLine = VersionLine{
				version, fields[1], fields[2], recordsCount,
				fields[4], fields[5], fields[6],
			}
		} else if strings.HasSuffix(line, "summary") {
			count, _ := strconv.Atoi(fields[4])
			summary := Summary{fields[0], fields[2], count}
			summaries = append(summaries, summary)
		} else {
			count, _ := strconv.Atoi(fields[4])
			if strings.HasPrefix(fields[2], "ipv") {
				record := IpRecord{
					&Record{fields[0], fields[1], fields[2],
						count, fields[5], fields[6]},
					net.ParseIP(fields[3]),
				}
				ipRecords = append(ipRecords, record)
			} else if strings.HasPrefix(fields[2], "asn") {
				asnNumber, _ := strconv.Atoi(fields[3])
				record := AsnRecord{
					&Record{fields[0], fields[1], fields[2],
						count, fields[5], fields[6]},
					asnNumber,
				}
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
