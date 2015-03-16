package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// Fri Feb 27 22:11:38 CET 2015 File of 4.2M
// BenchmarkReader	       1	1181852831 ns/op (~1.18s)

// Mon Mar 16 14:33:52 CET 2015
// BenchmarkReader	       1	1235675316 ns/op
func BenchmarkReader(b *testing.B) {
	path := filepath.Join(os.Getenv("HOME"), ".rir", "ripencc", "latest")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Cannot read file for bench. %v", err)
	}
	if len(content) == 0 {
		log.Fatal(" File for bench is empty!")
	}
	b.ResetTimer()
	data := bytes.NewBuffer(content)
	NewReader(data).Read()
}

func findIpWith(records *Records, address string) *IpRecord {
	for _, ip := range records.Ips {
		if address == ip.Start.String() {
			return ip
		}
	}
	log.Fatalf("Cannot find ip with address %s", address)
	return &IpRecord{}
}

func findAsnWith(records *Records, number int) *AsnRecord {
	for _, asn := range records.Asns {
		if number == asn.Start {
			return asn
		}
	}
	log.Fatal("Cannot find asn with number %s", number)
	return &AsnRecord{}

}

func TestParsingRegularFile(t *testing.T) {
	data := bytes.NewBufferString(regularData)

	records, _ := NewReader(data).Read()

	recordsCount, asnCount, ipv4Count, ipv6Count := 23486, 3986, 17947, 1553

	if records.Version != 2.3 {
		t.Errorf("records version: expected 2.3 got %d", records.Version)
	}
	if records.Count != recordsCount {
		t.Errorf("total records count: expected %d got %d", recordsCount, records.Count)
	}
	if records.AsnCount != asnCount {
		t.Errorf("asn count: expected %d got %d", asnCount, records.AsnCount)
	}
	if records.Ipv4Count != ipv4Count {
		t.Errorf("ipv4 count: expected %d got %d", ipv4Count, records.Ipv4Count)
	}
	if records.Ipv6Count != ipv6Count {
		t.Errorf("ipv6 count: expected %d got %d", ipv6Count, records.Ipv6Count)
	}

	if len(records.Asns) != 2 {
		t.Errorf("asn real count: expected %d got %d", 2, len(records.Asns))
	}

	if len(records.Ips) != 9 {
		t.Errorf("ips real count: expected %d got %d. Content:", 9, len(records.Ips), records.Ips)
	}

	asnRecord := findAsnWith(records, 173)
	if asnRecord.Status != "allocated" {
		t.Errorf("asn record status: expected 'allocated' got %q", asnRecord.Status)
	}

	ipRecord := findIpWith(records, "203.81.160.0")
	if ipRecord.Status != "assigned" {
		t.Errorf("ip record status: expected 'assigned' got %q", ipRecord.Status)
	}

	otherIpRecord := findIpWith(records, "193.9.26.0")
	if otherIpRecord.Status != "assigned" {
		t.Errorf("ip record status: expected 'assigned' got %q", otherIpRecord.Status)
	}

}

var regularData = `2.3|apnic|20110113|23486|19850701|20110112|+1000
# line to be ignored
apnic|*|asn|*|3986|summary
apnic|*|ipv4|*|17947|summary

apnic|*|ipv6|*|1553|summary
apnic|JP|asn|173|1|20020801|allocated
apnic|NZ|asn|681|1|20020801|allocated
apnic|MM|ipv4|203.81.64.0|8192|20100504|assigned
apnic|MM|ipv4|203.81.160.0|4096|20100122|assigned
apnic|KP|ipv4|175.45.176.0|1024|20100122|assigned
apnic|JP|ipv6|2001:200::|35|19990813|allocated
apnic|JP|ipv6|2001:200:2000::|35|20030423|allocated
apnic|JP|ipv6|2001:200:4000::|34|20030423|allocated
apnic|JP|ipv6|2001:200:8000::|33|20030423|allocated
ripencc|PL|ipv4|193.9.25.0|256|20090225|assigned
ripencc|HU|ipv4|193.9.26.0|512|20081222|assigned`
