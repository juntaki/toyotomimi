package radiolib

import "testing"

func TestRadiruStation(t *testing.T) {
	s := GetRadiruStations()
	if len(s) == 0 {
		t.Fatal(len(s))
	}

	r := NewRecorder(s[0], "/tmp")
	r.debug = true
	r.Record()
}
