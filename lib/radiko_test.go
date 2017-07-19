package radiolib

import (
	"sync"
	"testing"
	"time"

	radiko "github.com/yyoshiki41/go-radiko"
)

type stubRadikoClient struct{}

func (*stubRadikoClient) AuthToken() string                     { return "token" }
func (*stubRadikoClient) authorize()                            {}
func (*stubRadikoClient) GetStations() (radiko.Stations, error) { return radiko.Stations{}, nil }

func TestRadikoRefresh(t *testing.T) {
	// Stub authorize()
	refreshMutex := &sync.Mutex{}
	refreshedTime := new(time.Time)
	*refreshedTime = time.Now().Add(-100 * time.Minute)
	client := &stubRadikoClient{}

	rs1 := &radikoStation{
		client:        client,
		refreshMutex:  refreshMutex,
		refreshedTime: refreshedTime,
	}
	rs2 := &radikoStation{
		client:        client,
		refreshMutex:  refreshMutex,
		refreshedTime: refreshedTime,
	}

	rs1.Refresh()
	rs2.Refresh()
}

func TestRadikoStation(t *testing.T) {
	s := GetRadikoStations()
	if len(s) == 0 {
		t.Fatal(len(s))
	}

	r := NewRecorder(s[0], "/tmp")
	r.debug = true
	r.Record()

	p := s[0].NextProgram()
	if (time.Now()).After(p.end) {
		t.Fatal(p.end)
	}

	if s[0].Name() == "" {
		t.Fatal(s[0].Name())
	}
}
