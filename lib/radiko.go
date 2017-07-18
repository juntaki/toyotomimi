package radiolib

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"path"
	"time"

	radiko "github.com/yyoshiki41/go-radiko"
)

const radikoPlayerURL = "http://radiko.jp/apps/js/flash/myplayer-release.swf"

func GetRadikoStations() []Station {
	client, err := radiko.New("")
	if err != nil {
		panic(err)
	}

	authorize(client)

	stations, err := client.GetStations(context.Background(), time.Now())
	if err != nil {
		log.Fatal(err)
	}

	ret := make([]Station, len(stations))

	for i, s := range stations {
		ret[i] = &radikoStation{
			client:    client,
			station:   s,
			nextIndex: 0,
			streamURL: []radiko.URLItem{},
		}
	}

	return ret
}

func authorize(client *radiko.Client) {
	dir := "/tmp/"

	swfPath := path.Join(dir, "myplayer.swf")
	if err := radiko.DownloadPlayer(swfPath); err != nil {
		log.Fatalf("Failed to download swf player. %s", err)
	}

	cmdPath, err := exec.LookPath("swfextract")
	if err != nil {
		log.Fatal(err)
	}

	authKeyPath := path.Join(dir, "authkey.png")
	if c := exec.Command(cmdPath, "-b", "12", swfPath, "-o", authKeyPath); err != c.Run() {
		log.Fatalf("Failed to execute swfextract. %s", err)
	}

	_, err = client.AuthorizeToken(context.Background(), authKeyPath)
	if err != nil {
		log.Fatal(err)
	}
}

type radikoStation struct {
	client    *radiko.Client // Use the same by all stations
	station   radiko.Station
	nextIndex int
	streamURL []radiko.URLItem // for cache
}

func (r *radikoStation) getStation() {
	stations, err := r.client.GetStations(context.Background(), time.Now())
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range stations {
		if r.station.Name == s.Name {
			r.station = s
		}
	}
}

func (r *radikoStation) NextProgram() program {
	programs := r.station.Progs.Progs
	now := time.Now()

	for {
		if len(programs) == r.nextIndex {
			r.getStation()
			r.nextIndex = 0
		}
		p := programs[r.nextIndex]
		r.nextIndex++

		loc, _ := time.LoadLocation("Asia/Tokyo")

		start, _ := time.ParseInLocation("20060102150405", p.Ft, loc)
		end, _ := time.ParseInLocation("20060102150405", p.To, loc)

		if now.After(end) {
			continue
		}

		return program{
			title: p.Title,
			start: start,
			end:   end,
		}
	}
}

func (r *radikoStation) Refresh() {
	logger.Info("Token refresh: ", r.client.AuthToken())
	authorize(r.client)
	logger.Info("NewToken: ", r.client.AuthToken())
}

func (r *radikoStation) Name() string {
	return r.station.Name
}

func (r *radikoStation) URL() string {
	if len(r.streamURL) == 0 {
		r.streamURL, _ = radiko.GetStreamMultiURL(r.station.ID)
	}
	url := fmt.Sprintf("%s swfUrl=%s swfVfy=1 conn=S: conn=S: conn=S: conn=S:%s live=1 timeout=10",
		r.streamURL[rand.Intn(len(r.streamURL))].Item, radikoPlayerURL, r.client.AuthToken())
	return url
}
