package radiolib

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const radiruPlayerURL = "http://www3.nhk.or.jp/netradio/files/swf/rtmpe_ver2015.swf"
const radiruProgramURL = "http://www2.nhk.or.jp/hensei/api/sche.cgi"
const radiruConfigURL = "http://www3.nhk.or.jp/netradio/app/config_pc_2016.xml"

var radiruStations = []string{"r1", "r2", "fm"}

type RadiruProgram struct {
	StartTime string `xml:"starttime"`
	EndTime   string `xml:"endtime"`
	Title     string `xml:"title"`
}

type RadiruList struct {
	Item []RadiruProgram `xml:"item>item"`
}

func getStation(stationID string) []RadiruProgram {
	// Yesterday
	progURL := fmt.Sprintf("%s?c=4&mode=xml&ch=net%s&date=%s&tz=all",
		radiruProgramURL, stationID, time.Now().Add(-24*time.Hour).Format("20060102"))

	pYesterday := &RadiruList{}
	FetchXML(progURL, pYesterday)

	// Today
	progURL = fmt.Sprintf("%s?c=4&mode=xml&ch=net%s&date=%s&tz=all",
		radiruProgramURL, stationID, time.Now().Format("20060102"))
	pToday := &RadiruList{}
	FetchXML(progURL, pToday)

	// Concat
	programs := append(pYesterday.Item, pToday.Item...)

	return programs
}

type RadiruConfig struct {
	AreaKey string `xml:"areakey"`
	R1      string `xml:"r1"`
	R2      string `xml:"r2"`
	FM      string `xml:"fm"`
}

func (r RadiruConfig) Get(stationID string) string {
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

type RadiruConfigList struct {
	Data []RadiruConfig `xml:"stream_url>data"`
}

func getConfig(areaKey string) RadiruConfig {
	list := &RadiruConfigList{}
	FetchXML(radiruConfigURL, list)

	for _, c := range list.Data {
		if c.AreaKey == areaKey {
			return c
		}
	}
	log.Fatal("Invalid AreaKey")
	return RadiruConfig{}
}

func GetRadiruStations() []Station {
	ret := make([]Station, len(radiruStations))

	config := getConfig("130") // TODO: Autodetect

	for i, s := range radiruStations {
		ret[i] = &RadiruStation{
			stationID: s,
			nextIndex: 0,
			configURL: config.Get(s),
			program:   getStation(s),
		}
	}
	return ret
}

type RadiruStation struct {
	stationID string // r1 r2 fm
	nextIndex int
	configURL string
	program   []RadiruProgram
}

func (r *RadiruStation) NextProgram() Program {
	programs := r.program
	now := time.Now()
	for {
		if len(programs) == r.nextIndex {
			r.program = getStation(r.stationID)
			r.nextIndex = 0
		}
		p := programs[r.nextIndex]
		r.nextIndex++

		loc, _ := time.LoadLocation("Asia/Tokyo")

		start, _ := time.ParseInLocation("2006-01-02 15:04:05", p.StartTime, loc)
		end, _ := time.ParseInLocation("2006-01-02 15:04:05", p.EndTime, loc)

		if now.After(end) {
			continue
		}

		return Program{
			url:   r.url(),
			title: p.Title,
			start: start,
			end:   end,
		}
	}
}

func (r *RadiruStation) StationName() string {
	switch r.stationID {
	case "r1":
		return "ラジオ第1"
	case "r2":
		return "ラジオ第2"
	case "fm":
		return "NHK-FM"
	}
	return ""
}

func (r *RadiruStation) url() string {
	url := fmt.Sprintf("%s swfUrl=%s swfVfy=1 live=1",
		r.configURL, radiruPlayerURL)
	return url
}

func FetchXML(url string, v interface{}) {
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
