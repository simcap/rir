package cache

var (
	Providers = map[string]string{
		"afrinic": "http://ftp.apnic.net/stats/afrinic/delegated-afrinic-latest",
		"apnic":   "http://ftp.apnic.net/stats/apnic/delegated-apnic-latest",
		"iana":    "http://ftp.apnic.net/stats/iana/delegated-iana-latest",
		"lacnic":  "http://ftp.apnic.net/stats/lacnic/delegated-lacnic-latest",
		"ripencc": "http://ftp.apnic.net/stats/ripe-ncc/delegated-ripencc-latest",
	}
)
