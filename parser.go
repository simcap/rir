package rir

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type Summary struct {
	Registry string
	Type     string
	Count    int64
}

type Record struct {
	Registry string
	Cc       string
	Type     string
	Start    string
	Value    int64
	Date     string
	Status   string
}

type RirData struct {
	AsnCount  int64
	Ipv4Count int64
	Ipv6Count int64
}

func Parse(r io.Reader) *RirData {
	records := []Record{}
	summaries := []Summary{}

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		} else if strings.HasSuffix(line, "summary") {
			fields := strings.Split(line, "|")
			count, _ := strconv.ParseInt(fields[4], 10, 0)
			summary := Summary{fields[0], fields[2], count}
			summaries = append(summaries, summary)
		} else {
			fields := strings.Split(line, "|")
			count, _ := strconv.ParseInt(fields[4], 10, 0)
			record := Record{fields[0], fields[1], fields[2], fields[3], count, fields[5], fields[6]}
			records = append(records, record)
		}
	}

	var asnCount, ipv4Count, ipv6Count int64
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

	return &(RirData{AsnCount: asnCount, Ipv4Count: ipv4Count, Ipv6Count: ipv6Count})
}
