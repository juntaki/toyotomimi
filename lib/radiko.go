package radiolib

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"path"
	"sync"
	"time"

	radiko "github.com/yyoshiki41/go-radiko"
)

const radikoPlayerURL = "http://radiko.jp/apps/js/flash/myplayer-release.swf"

func GetRadikoStations() []Station {
	client := newRadikoClient()

	refreshMutex := &sync.Mutex{}
	refreshedTime := new(time.Time)
	*refreshedTime = time.Now()

	client.authorize()

	stations, err := client.GetStations()
	if err != nil {
		log.Fatal(err)
	}

	ret := make([]Station, len(stations))

	for i, s := range stations {
		ret[i] = &radikoStation{
			client:        client,
			station:       s,
			nextIndex:     0,
			streamURL:     []radiko.URLItem{},
			refreshMutex:  refreshMutex,
			refreshedTime: refreshedTime,
		}
	}

	return ret
}

type radikoClientInterface interface {
	GetStations() (radiko.Stations, error)
	AuthToken() string
	authorize()
}

// Wrap for stubbing
type radikoClient struct {
	client *radiko.Client
}

func newRadikoClient() *radikoClient {
	client, err := radiko.New("")
	if err != nil {
		panic(err)
	}

	return &radikoClient{
		client: client,
	}
}

func (rc *radikoClient) GetStations() (radiko.Stations, error) {
	return rc.client.GetStations(context.Background(), time.Now())
}

func (rc *radikoClient) AuthToken() string {
	return rc.client.AuthToken()
}

func (rc *radikoClient) authorize() {
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

	_, err = rc.client.AuthorizeToken(context.Background(), authKeyPath)
	if err != nil {
		log.Fatal(err)
	}
}

type radikoStation struct {
	client        radikoClientInterface // Use the same by all stations
	station       radiko.Station
	nextIndex     int
	streamURL     []radiko.URLItem // for cache
	refreshMutex  *sync.Mutex
	refreshedTime *time.Time
}

func (r *radikoStation) getStation() {
	stations, err := r.client.GetStations()
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
	r.refreshMutex.Lock()
	if r.refreshedTime.Add(10 * time.Minute).Before(time.Now()) {
		logger.Info("Token refresh (old): ", r.client.AuthToken())
		r.client.authorize()
		*r.refreshedTime = time.Now()
		logger.Info("Token refresh (new): ", r.client.AuthToken())
	} else {
		logger.Info("Token not refresh")
	}

	r.refreshMutex.Unlock()
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
