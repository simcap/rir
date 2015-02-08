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

	rirData := rir.Parse(file)

	asnCount, ipv4Count, ipv6Count := 3986, 17947, 1553

	if rirData.AsnCount != asnCount {
		t.Errorf("asn count: expected %d got %d", asnCount, rirData.AsnCount)
	}
	if rirData.Ipv4Count != ipv4Count {
		t.Errorf("ipv4 count: expected %d got %d", ipv4Count, rirData.Ipv4Count)
	}
	if rirData.Ipv6Count != ipv6Count {
		t.Errorf("ipv6 count: expected %d got %d", ipv6Count, rirData.Ipv6Count)
	}

}
