package radiolib

import "testing"

func TestRadiruStation(t *testing.T) {
	s := GetRadiruStations()
	if len(s) == 0 {
		t.Fatal(len(s))
	}

	r := NewRecorder(s[0], "/home/juntaki/")
	r.debug = true
	r.Record()
}
