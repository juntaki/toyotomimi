package radiolib

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path"
	"time"

	radiko "github.com/yyoshiki41/go-radiko"
)

const playerURL = "http://radiko.jp/apps/js/flash/myplayer-release.swf"

func getRadikoStations() []Station {
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
		ret[i] = &RadikoStation{
			client:    client,
			station:   s,
			nextIndex: 0,
			streamURL: "",
		}
	}

	return ret
}

func authorize(client *radiko.Client) {
	dir := "/tmp/"

	// 1. Download a swf player.
	swfPath := path.Join(dir, "myplayer.swf")
	if err := radiko.DownloadPlayer(swfPath); err != nil {
		log.Fatalf("Failed to download swf player. %s", err)
	}

	// 2. Using swfextract, create an authkey file from a swf player.
	cmdPath, err := exec.LookPath("swfextract")
	if err != nil {
		log.Fatal(err)
	}
	authKeyPath := path.Join(dir, "authkey.png")
	if c := exec.Command(cmdPath, "-b", "12", swfPath, "-o", authKeyPath); err != c.Run() {
		log.Fatalf("Failed to execute swfextract. %s", err)
	}

	// 4. Enables and sets the auth_token.
	// After client.AuthorizeToken() has succeeded,
	// the client has the enabled auth_token internally.
	_, err = client.AuthorizeToken(context.Background(), authKeyPath)
	if err != nil {
		log.Fatal(err)
	}
}

type RadikoStation struct {
	client    *radiko.Client // Use the same by all stations
	station   radiko.Station
	nextIndex int
	streamURL string // for cache
}

func (r *RadikoStation) getStation() {
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

func (r *RadikoStation) NextProgram() Program {
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

		return Program{
			url:   r.url(),
			title: p.Title,
			start: start,
			end:   end,
		}
	}
}

func (r *RadikoStation) StationName() string {
	return r.station.Name
}

func (r *RadikoStation) url() string {
	if r.streamURL == "" {
		items, _ := radiko.GetStreamMultiURL(r.station.ID)
		r.streamURL = items[0].Item
	}

	url := fmt.Sprintf("%s swfUrl=%s swfVfy=1 conn=S: conn=S: conn=S: conn=S:%s live=1 timeout=10",
		r.streamURL, playerURL, r.client.AuthToken())
	return url
}
