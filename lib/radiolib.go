package radiolib

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
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

	rlogger := logger.WithFields(logrus.Fields{
		"start": start,
		"path":  targetPath,
	})

	file, err := os.Create(targetPath)
	if err != nil {
		rlogger.Fatal("Create", err)
	}
	defer file.Close()

	if p.start.After(time.Now()) {
		time.Sleep(time.Until(p.start))
		rlogger.Info("Wait until program start")
	}

	b := make([]byte, 1024)
	retry := 0

reconnect:
	rlogger.Info("Start Recording")
	r, _ := rtmp.Init()
	r.SetupURL(p.url)
	r.Connect()
	defer r.Close()

	for {
		size, err := r.Read(b)
		if size <= 0 || err != nil {
			if retry > 3 {
				rlogger.Error("Read failed", err)
			}
			rlogger.Error("Read failed, try reconnect", err)
			r.Close()
			retry++
			goto reconnect
		}

		wsize, err := file.Write(b[:size])
		if size != wsize || err != nil {
			rlogger.Error("Write failed, just ignore", err)
		}

		if time.Now().After(p.end) {
			rlogger.Info("End Recording")
			break
		}

		if rec.debug && time.Now().After(start.Add(10*time.Second)) {
			break
		}
	}

	if p.end.After(time.Now()) {
		time.Sleep(time.Until(p.end))
		rlogger.Error("Wait until program end, due to some error")
	}
}

func RecordAll(outputDir string) {
	var wg sync.WaitGroup
	radiko := getRadikoStations()
	radiru := GetRadiruStations()
	stations := append(radiko, radiru...)
	for _, s := range stations {
		rec := NewRecorder(s, outputDir)
		wg.Add(1)
		go func(rec *Recorder) {
			defer wg.Done()
			for {
				rec.Record()
			}
		}(rec)
		time.Sleep(time.Second)
	}
	wg.Wait()
}
