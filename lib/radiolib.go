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
	Name() string
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
	recordStart := time.Now()
	prog := rec.station.NextProgram()

	filename := fmt.Sprintf("[%s][%s]%s.m4a",
		prog.start.Format("2006-0102-1504"), rec.station.Name(),
		strings.Replace(prog.title, "/", "_", -1)) // "/" is filepath separator
	rlogger := logger.WithFields(logrus.Fields{"filename": filename})

	filepath := path.Join(rec.outputDir, filename)

	rlogger.Debug("filepath: ", filepath)
	rlogger.Debug("recordStart:", recordStart)
	rlogger.Debug("program:", prog)

	file, err := os.Create(filepath)
	if err != nil {
		rlogger.Error("Create failed, skipped. err=", err)
		return
	}
	defer file.Close()

	if prog.start.After(time.Now()) {
		rlogger.Info("Wait until program start. until=", time.Until(prog.start))
		time.Sleep(time.Until(prog.start))
	}

	buf := make([]byte, 64*1024)
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

		err = r.SetupURL(prog.url)
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
			size, err := r.Read(buf)
			if size <= 0 || err != nil {
				rlogger.Error("Read failed, try reconnect", err)
				r.Close()
				continue connect
			}

			wsize, err := file.Write(buf[:size])
			if size != wsize || err != nil {
				rlogger.Error("Write failed, just ignore", err)
			}

			if time.Now().After(prog.end) {
				rlogger.Info("End Recording")
				break connect
			}

			if rec.debug && time.Now().After(recordStart.Add(time.Second)) {
				rlogger.Info("End Recording (debug)")
				break connect
			}
		}
	}
	r.Close()
}

func RecordAll(outputDir string) {
	var wg sync.WaitGroup
	radiko := GetRadikoStations()
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
