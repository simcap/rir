package reader

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/simcap/rir/reader"
)

// Fri Feb 27 22:11:38 CET 2015
// BenchmarkReader	       1	1181852831 ns/op (~1.18s)
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
	reader.NewReader(bytes.NewBuffer(content)).Read()
}

func TestParsingRegularFile(t *testing.T) {
	data := bytes.NewBufferString(regularData)

	records, _ := reader.NewReader(data).Read()

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

	firstAsnRecord := records.Asns[0]
	if firstAsnRecord.Status != "allocated" {
		t.Errorf("asn record status: expected 'allocated' got %q", firstAsnRecord.Status)
	}
	if firstAsnRecord.Start != 173 {
		t.Errorf("asn record status: expected 173 got %q", firstAsnRecord.Start)
	}

	firstIpRecord := records.Ips[1]
	if firstIpRecord.Status != "assigned" {
		t.Errorf("asn record status: expected 'assigned' got %q", firstIpRecord.Status)
	}
	if firstIpRecord.Start.String() != "203.81.160.0" {
		t.Errorf("asn record status: expected '203.81.160.0' got %q", firstIpRecord.Start.String())
	}

}

func TestRaiseErrors(t *testing.T) {
	data := bytes.NewBufferString(faultyData)

	_, err := reader.NewReader(data).Read()
	perr, _ := err.(*reader.ParseError)

	if err == nil {
		t.Error("expecting an error to occur")
	} else if perr.Line != 5 {
		t.Errorf("error line: expecting line 5 got %d", perr.Line)
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

var faultyData = `2.3|apnic|20110113|23486|19850701|20110112|+1000
apnic|*|asn|*|3986|summary
apnic|*|ipv4|*|17947|summary
apnic|*|ipv6|*|1553|summary
2.3|apnic|20110113|23486|19850701|20110112|+1000`
