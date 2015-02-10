package rir_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/simcap/rir"
)

func TestParsing(t *testing.T) {
	file, err := os.Open("./data/rir.txt")
	defer file.Close()

	if err != nil {
		fmt.Fprintln(os.Stdout, "cannot find file")
	}

	records := rir.Parse(file)

	recordsCount, asnCount, ipv4Count, ipv6Count := 23486, 3986, 17947, 1553

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

}
