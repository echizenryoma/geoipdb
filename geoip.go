package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"sort"
)

type CountryStatisticsTable []uint16

type GeoIPv4 struct {
	CountryCode uint16
	BeginSubnet uint32
	EndSubnet   uint32
}

type GeoIPv4Table []GeoIPv4

func (table GeoIPv4Table) Len() int      { return len(table) }
func (table GeoIPv4Table) Swap(i, j int) { table[i], table[j] = table[j], table[i] }
func (table GeoIPv4Table) Less(i, j int) bool {
	return table[i].CountryCode < table[j].CountryCode ||
		table[i].CountryCode == table[j].CountryCode && table[i].BeginSubnet < table[j].BeginSubnet
}

func (table GeoIPv4Table) CountByCountry() (statistics CountryStatisticsTable) {
	if len(table) == 0 {
		return
	}
	countryIndex := 0
	statistics = append(statistics, uint16(0))
	currentCountryCode := table[0].CountryCode
	for _, geoipv4 := range table {
		if currentCountryCode == geoipv4.CountryCode {
			statistics[countryIndex]++
		} else {
			statistics = append(statistics, uint16(1))
			countryIndex++
			currentCountryCode = geoipv4.CountryCode
		}
	}
	return
}

func WriteGeoIPv4Database(binFileName string, idxFileName string, geoipv4Table GeoIPv4Table, countryStatisticsTable CountryStatisticsTable) (err error) {
	if len(geoipv4Table) == 0 {
		err = errors.New("GeoIPv4Table is nil")
		return
	}

	binFile, err := os.Create(binFileName)
	if err != nil {
		return
	}
	defer binFile.Close()

	idxFile, err := os.Create(idxFileName)
	if err != nil {
		return
	}
	defer idxFile.Close()

	countryIndex := 0
	currentCountryCode := uint16(0)
	binOffset := uint32(0)

	/*
		GeoIP Database Format:
		+----------------+----------------+
		|  Country Code  |     Number     |
		|    (2 Bytes)   |    (2 Bytes)   |
		+----------------+--------------------^---+
		|           Begin Subnet          |   |
		|            (4 Bytes)            |   +
		+---------------------------------+ Number
		|           End Subnet            |   +
		|            (4 Bytes)            |   |
		+---------------------------------+---v---+


		GeoIP Indx file Format:
		+-------------------------------------------------+
		|  Country Code  |       Offset in Database       |
		|    (2 Bytes)   |             (4 Bytes)          |
		+-------------------------------------------------+
	*/

	var binBuff bytes.Buffer
	var idxBuff bytes.Buffer

	for _, geoipv4 := range geoipv4Table {
		if currentCountryCode != geoipv4.CountryCode {
			currentCountryCode = geoipv4.CountryCode

			binary.Write(&idxBuff, binary.LittleEndian, currentCountryCode)
			binary.Write(&idxBuff, binary.LittleEndian, binOffset)

			binary.Write(&binBuff, binary.LittleEndian, currentCountryCode)
			binary.Write(&binBuff, binary.LittleEndian, countryStatisticsTable[countryIndex])
			binOffset += 4

			countryIndex++
		}

		binary.Write(&binBuff, binary.LittleEndian, geoipv4.BeginSubnet)
		binary.Write(&binBuff, binary.LittleEndian, geoipv4.EndSubnet)

		binOffset += 8
	}

	idxFile.Write(idxBuff.Bytes())
	binFile.Write(binBuff.Bytes())
	return
}

func Csv2Bin(geoipv4Table GeoIPv4Table, binFileName string, idxFileName string) (err error) {
	sort.Sort(geoipv4Table)
	countryStatisticsTable := geoipv4Table.CountByCountry()
	err = WriteGeoIPv4Database(binFileName, idxFileName, geoipv4Table, countryStatisticsTable)
	return
}
