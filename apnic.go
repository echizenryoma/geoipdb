package main

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type ApnicVersion struct {
	Version   string
	Registry  string
	Serial    string
	Records   uint32
	StartDate string
	EndDate   string
	UTCOffset string
}

func (version ApnicVersion) parse(recordStr string) (err error) {
	cols := strings.Split(recordStr, "|")
	if len(cols) < reflect.TypeOf(version).NumField() {
		err = errors.New("version string is too short")
		return
	}

	version.Version = cols[0]
	version.Registry = cols[1]
	version.Serial = cols[2]
	num, err := strconv.Atoi(cols[3])
	if err != nil {
		return
	}
	version.Records = uint32(num)
	version.StartDate = cols[4]
	version.EndDate = cols[5]
	version.UTCOffset = cols[6]
	return
}

type ApnicSummary struct {
	Registry string
	Type     string
	Count    uint32
	Summary  string
}

func (this *ApnicSummary) parse(recordStr string) (err error) {
	cols := strings.Split(recordStr, "|")
	if len(cols) < reflect.TypeOf(*this).NumField()+2 {
		err = errors.New("summary string is too short")
		return
	}
	this.Registry = cols[0]
	this.Type = cols[2]
	num, err := strconv.Atoi(cols[4])
	if err != nil {
		return
	}
	this.Count = uint32(num)
	this.Summary = cols[5]
	return
}

type ApnicSummaryArray []ApnicSummary

type ApnicRecord struct {
	Registry string
	CC       string
	Type     string
	Start    string
	Value    uint32
	Date     string
	Status   string
}

func (this *ApnicRecord) parse(recordStr string) (err error) {
	cols := strings.Split(recordStr, "|")
	if len(cols) < reflect.TypeOf(*this).NumField() {
		err = errors.New("record string is too short")
		return
	}
	this.Registry = cols[0]
	this.CC = cols[1]
	this.Type = cols[2]
	this.Start = cols[3]
	num, err := strconv.Atoi(cols[4])
	if err != nil {
		return
	}
	this.Value = uint32(num)
	this.Date = cols[5]
	this.Status = cols[6]
	return
}

type ApnicRecordArray []ApnicRecord

func (array ApnicRecordArray) SelectCountryCode(countryCode string) (recordsDst ApnicRecordArray, err error) {
	for _, record := range array {
		if record.CC == countryCode {
			recordsDst = append(recordsDst, record)
		}
	}
	return
}

func (array ApnicRecordArray) SelectType(typeStr string) (recordsDst ApnicRecordArray, err error) {
	for _, record := range array {
		if record.Type == typeStr {
			recordsDst = append(recordsDst, record)
		}
	}
	return
}

func (array ApnicRecordArray) ToGeoIPv4Table() (geoipv4Array GeoIPv4Table, err error) {
	for _, record := range array {
		if record.Type == "ipv4" {
			beginIP := IP2Uint32(net.ParseIP(record.Start))
			hostNum := record.Value
			endIP := beginIP + hostNum
			countryCode := binary.BigEndian.Uint16([]byte(record.CC))

			geoipv4Array = append(geoipv4Array, GeoIPv4{CountryCode: countryCode, BeginSubnet: beginIP, EndSubnet: endIP})
		}
	}
	return
}

func ParseRecords(url string) (version ApnicVersion, summarys ApnicSummaryArray, records ApnicRecordArray, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	lines := strings.Split(string(body), "\n")

	index := 0
	for index < len(lines) && (lines[index] == "" || strings.HasPrefix(lines[index], "#")) {
		index++
	}

	line := lines[index]
	version.parse(line)
	index++

	for index < len(lines) {
		line = lines[index]
		if len(line) == 0 {
			index++
			continue
		}

		if !strings.HasSuffix(lines[index], "summary") {
			break
		}

		var summary ApnicSummary
		err = summary.parse(line)
		if err != nil {
			return
		}
		summarys = append(summarys, summary)
		index++
	}

	for index < len(lines) {
		line = lines[index]
		if len(line) == 0 {
			index++
			continue
		}

		var record ApnicRecord
		err = record.parse(line)
		if err != nil {
			return
		}
		records = append(records, record)
		index++
	}
	return
}
