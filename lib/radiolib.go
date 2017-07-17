package radiolib

import (
	"fmt"
	"os"
	"path"
	"strings"
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
		p.start.Format("2006-0102-1504"), rec.station.StationName(),
		strings.Replace(p.title, "/", "_", -1))
	targetPath := path.Join(rec.outputDir, filename)

	rlogger := logger.WithFields(logrus.Fields{
		"start": start,
		"path":  targetPath,
	})

	file, err := os.Create(targetPath)
	if err != nil {
		rlogger.Error("Create", err)
		return
	}
	defer file.Close()

	if p.start.After(time.Now()) {
		rlogger.Info("Wait until program start", time.Until(p.start))
		time.Sleep(time.Until(p.start))
	}

	b := make([]byte, 64*1024)
	connect := 0

	r, err := rtmp.Alloc()
	defer r.Free()
	r.Init()
connect:
	for {
		connect++
		if connect > 4 {
			rlogger.Error("Too much retry.")
			return
		}

		err = r.SetupURL(p.url)
		if err != nil {
			rlogger.Error("SetupURL failed", err)
			continue connect
		}
		rlogger.Info("Start Recording")
		err = r.Connect()
		if err != nil {
			rlogger.Error("Connect failed", err)
			continue connect
		}

		for {
			size, err := r.Read(b)
			if size <= 0 || err != nil {
				rlogger.Error("Read failed, try reconnect", err)
				r.Close()
				continue connect
			}

			wsize, err := file.Write(b[:size])
			if size != wsize || err != nil {
				rlogger.Error("Write failed, just ignore", err)
			}

			if time.Now().After(p.end) {
				rlogger.Info("End Recording")
				break connect
			}

			if rec.debug && time.Now().After(start.Add(time.Second)) {
				break connect
			}
		}
	}
	r.Close()
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
