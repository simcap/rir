package rir_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/simcap/netblocks/rir"
)

func TestParsing(t *testing.T) {
	file, err := os.Open("./data/rir.txt")
	defer file.Close()

	if err != nil {
		fmt.Fprintln(os.Stdout, "cannot find file")
	}

	rirData := rir.Parse(file)

	asnCount, ipv4Count, ipv6Count := int64(3986), int64(17947), int64(1553)

	if rirData.AsnCount != asnCount {
		t.Errorf("asn count: expected %q got %q", asnCount, rirData.AsnCount)
	}
	if rirData.Ipv4Count != ipv4Count {
		t.Errorf("ipv4 count: expected %q got %q", ipv4Count, rirData.Ipv4Count)
	}
	if rirData.Ipv6Count != ipv6Count {
		t.Errorf("ipv6 count: expected %q got %q", ipv6Count, rirData.Ipv6Count)
	}

}
