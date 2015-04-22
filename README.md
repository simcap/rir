# RIR files explorer

Go (golang) package to parse RIR files with command line utility

RIR stands for Regional Internet Registry. RIR files are used to exchange statistics around asn, ipv4 & ipv6. For more details on RIR files see [here](http://www.apnic.net/publications/media-library/documents/resource-guidelines/rir-statistics-exchange-format#FileHeader)

## Install

  1. Install Go beforehand
  2. Run `go get github.com/simcap/rir`

You should now have an executable `rir` in your path

## Test

Run `go test -v`

## Command line usage

Get the basic usage

    $ rir
    Usage of ./rir:
      -c="": 2 letters string of the country (ISO 3166)
      -q="": ip address to which to resolve country

Explore ip blocks given a country

    $ rir -c FR
    2.0.0.0/12
    5.10.128.0/21
    ...
    213.108.232.0/21
    213.111.0.0/18
    217.77.224.0/20

Get the country and IP net for an given IP

    $ rir -q 194.146.24.104
    FR 194.146.24.0/23

