package radiolib

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type radiruProgram struct {
	StartTime string `xml:"starttime"`
	EndTime   string `xml:"endtime"`
	Title     string `xml:"title"`
}

type radiruList struct {
	Item []radiruProgram `xml:"item>item"`
}

type radiruConfig struct {
	AreaKey string `xml:"areakey"`
	R1      string `xml:"r1"`
	R2      string `xml:"r2"`
	FM      string `xml:"fm"`
}

func (r radiruConfig) getStreamURL(stationID string) string {
	switch stationID {
	case "r1":
		return r.R1
	case "r2":
		return r.R2
	case "fm":
		return r.FM
	}
	return ""
}

type radiruConfigList struct {
	Data []radiruConfig `xml:"stream_url>data"`
}

func fetchXML(url string, v interface{}) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Fatal("Failed to fetch XML data")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err := xml.Unmarshal(body, v); err != nil {
		log.Fatal(err)
	}
}

func getProgram(stationID string) []radiruProgram {
	// Concat yesterday and today for midnight program
	progURL := fmt.Sprintf("%s?c=4&mode=xml&ch=net%s&date=%s&tz=all",
		radiruProgramURL, stationID, time.Now().Add(-24*time.Hour).Format("20060102"))
	pYesterday := &radiruList{}
	fetchXML(progURL, pYesterday)

	progURL = fmt.Sprintf("%s?c=4&mode=xml&ch=net%s&date=%s&tz=all",
		radiruProgramURL, stationID, time.Now().Format("20060102"))
	pToday := &radiruList{}
	fetchXML(progURL, pToday)

	programs := append(pYesterday.Item, pToday.Item...)

	return programs
}

func getConfig(areaKey string) radiruConfig {
	list := &radiruConfigList{}
	fetchXML(radiruConfigURL, list)

	for _, c := range list.Data {
		if c.AreaKey == areaKey {
			return c
		}
	}
	log.Fatal("Invalid AreaKey")
	return radiruConfig{}
}
