package radiolib

import (
	"fmt"
	"time"
)

const radiruPlayerURL = "http://www3.nhk.or.jp/netradio/files/swf/rtmpe_ver2015.swf"
const radiruProgramURL = "http://www2.nhk.or.jp/hensei/api/sche.cgi"
const radiruConfigURL = "http://www3.nhk.or.jp/netradio/app/config_pc_2016.xml"

var radiruStations = []string{"r1", "r2", "fm"}
var radiruStationsMap = map[string]string{
	"r1": "ラジオ第1",
	"r2": "ラジオ第2",
	"fm": "NHK-FM",
}

func GetRadiruStations() []Station {
	stations := make([]Station, len(radiruStations))

	config := getConfig("130") // TODO: Autodetect

	for i, stationID := range radiruStations {
		rStation := &radiruStation{
			stationID: stationID,
			nextIndex: 0,
			streamURL: config.getStreamURL(stationID),
		}
		rStation.setProgram()
		stations[i] = rStation
	}
	return stations
}

type radiruStation struct {
	stationID string // r1 r2 fm
	nextIndex int
	streamURL string
	program   []radiruProgram
}

func (r *radiruStation) setProgram() {
	r.program = getProgram(r.stationID)
}

func (r *radiruStation) NextProgram() program {
	programs := r.program
	now := time.Now()
	for {
		if len(programs) == r.nextIndex {
			r.setProgram()
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

		return program{
			url:   r.url(),
			title: p.Title,
			start: start,
			end:   end,
		}
	}
}

func (r *radiruStation) Name() string {
	return radiruStationsMap[r.stationID]
}

func (r *radiruStation) url() string {
	url := fmt.Sprintf("%s swfUrl=%s swfVfy=1 live=1 timeout=10",
		r.streamURL, radiruPlayerURL)
	return url
}
