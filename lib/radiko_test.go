package radiolib

import (
	"testing"
	"time"
)

func TestRadikoStation(t *testing.T) {
	s := getRadikoStations()
	if len(s) == 0 {
		t.Fatal(len(s))
	}

	r := NewRecorder(s[0], "/home/juntaki/")
	r.debug = true
	r.Record()

	p := s[0].NextProgram()
	if (time.Now()).After(p.end) {
		t.Fatal(p.end)
	}

	if s[0].StationName() == "" {
		t.Fatal(s[0].StationName())
	}
}
