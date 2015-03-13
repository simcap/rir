package rir

import (
	"log"
	"math"
	"net"
)

const (
	IPv4 = "ipv4"
	IPv6 = "ipv6"
)

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

func (ipr *IpRecord) Net() *net.IPNet {
	var mask net.IPMask
	if ipr.Type == IPv4 {
		hostscount := ipr.Value
		ones := 32 - int(math.Log2(float64(hostscount)))
		mask = net.CIDRMask(ones, 32)
	} else if ipr.Type == IPv6 {
		mask = net.CIDRMask(ipr.Value, 128)
	} else {
		log.Fatalf("no ipnet for ip of type '%s'", ipr.Type)
	}

	return &net.IPNet{ipr.Start, mask}
}
