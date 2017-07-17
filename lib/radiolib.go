package radiolib

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	rtmp "github.com/juntaki/go-librtmp"
)

type Station interface {
	NextProgram() Program
	StationName() string
}

type Program struct {
	url   string
	title string
	start time.Time
	end   time.Time
}

type Recorder struct {
	station   Station
	outputDir string
	debug     bool
}

func NewRecorder(station Station, outputDir string) *Recorder {
	return &Recorder{
		station:   station,
		outputDir: outputDir,
	}
}

func (rec *Recorder) Record() {
	start := time.Now()

	p := rec.station.NextProgram()
	filename := fmt.Sprintf("[%s][%s]%s.m4a",
		p.start.Format("2006-0102-1504"), rec.station.StationName(), p.title)
	targetPath := path.Join(rec.outputDir, filename)

	file, err := os.Create(targetPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// TODO: wait until start

	r, _ := rtmp.Init()
	r.SetupURL(p.url)
	r.Connect()
	defer r.Close()

	b := make([]byte, 1024)
	for {
		size, err := r.Read(b)
		if size <= 0 || err != nil {
			fmt.Println("read: ", size)
			fmt.Println(err)
			break
		}

		wsize, err := file.Write(b[:size])
		if size != wsize || err != nil {
			fmt.Println("write:", wsize, "!=", size)
			fmt.Println(err)
			break
		}

		if time.Now().After(p.end) {
			break
		}

		if rec.debug && time.Now().After(start.Add(10*time.Second)) {
			break
		}
	}
}

func RecordAll(outputDir string) {
	for _, s := range getRadikoStations() {
		rec := NewRecorder(s, outputDir)
		go func(rec *Recorder) {
			for {
				rec.Record()
			}
		}(rec)
	}
	for _, s := range GetRadiruStations() {
		rec := NewRecorder(s, outputDir)
		go func(rec *Recorder) {
			for {
				rec.Record()
			}
		}(rec)
	}
}
