package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	URL          string
	BinFileName  string
	IdxFileName  string
	CountryCodes string
)

func init() {
	log.SetOutput(os.Stdout)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nCopyright (c) 2018, Echizen Ryoma. All rights reserved.\n\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&URL, "url", "http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest", "Latest APNIC ASN Blocks URL")
	flag.StringVar(&BinFileName, "bin", "geoipdb.bin", "GeoIP Database File Path")
	flag.StringVar(&IdxFileName, "idx", "geoipdb.idx", "GeoIP Database Index File Path")
	flag.StringVar(&CountryCodes, "cc", "CN", "examples: \"CN\", \"CN, US\", \"All\"")
}

func main() {
	flag.Parse()

	CountryCodes = strings.ToUpper(CountryCodes)
	allFlag := strings.Contains(CountryCodes, "ALL")

	_, _, records, err := ParseRecords(URL)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	records, err = records.SelectType("ipv4")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	var SelectedRecords ApnicRecordArray
	if !allFlag {
		CountryCodes = strings.TrimSpace(CountryCodes)
		CountryCodes = strings.Replace(CountryCodes, " ", "", -1)
		for _, cc := range strings.Split(CountryCodes, ",") {
			selected, err := records.SelectCountryCode(cc)
			if err != nil {
				log.Fatalln(err.Error())
				return
			}
			SelectedRecords = append(SelectedRecords, selected...)
		}
	} else {
		SelectedRecords = records
	}

	geoipv4Table, err := SelectedRecords.ToGeoIPv4Table()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = Csv2Bin(geoipv4Table, BinFileName, IdxFileName)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
