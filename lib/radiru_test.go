package radiolib

import (
	"testing"
	"time"
)

func TestRadiruStation(t *testing.T) {
	s := GetRadiruStations()
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
