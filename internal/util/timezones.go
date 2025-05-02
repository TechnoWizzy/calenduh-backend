package util

import (
	"encoding/xml"
	"io"
	"log"
	"os"
)

type SupplementalData struct {
	WindowsZones WindowsZones `xml:"windowsZones"`
}

type WindowsZones struct {
	MapTimezones MapTimezones `xml:"mapTimezones"`
}

type MapTimezones struct {
	MapZones []MapZone `xml:"mapZone"`
}

type MapZone struct {
	Other     string `xml:"other,attr"`
	Territory string `xml:"territory,attr"`
	Type      string `xml:"type,attr"`
}

var Timezones map[string]string

func GetTimezone(tz string) string {
	return Timezones[tz]
}

func init() {
	xmlFile, err := os.Open("timezones.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer func(xmlFile *os.File) {
		err := xmlFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(xmlFile)

	data, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(err)
	}

	var supplementalData SupplementalData
	if err = xml.Unmarshal(data, &supplementalData); err != nil {
		log.Fatal(err)
	}

	Timezones = make(map[string]string)
	for _, zone := range supplementalData.WindowsZones.MapTimezones.MapZones {
		Timezones[zone.Other] = zone.Type
	}
}
