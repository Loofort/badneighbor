package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Loofort/badneighbor/ears"
	"github.com/Loofort/badneighbor/essentia"
	"github.com/Loofort/badneighbor/lame"
	"github.com/gordonklaus/portaudio"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var trace = flag.Bool("trace", false, "print additional info")
var level = flag.Float64("level", 0.002, "level of sound energy to record")

var (
	Objectives  = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	MaxAge      = 10 * time.Second
	soundEnergy = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sound_energy",
		Help:       "sound energy",
		Objectives: Objectives,
		MaxAge:     MaxAge,
	})
	soundLoudness = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sound_loudness",
		Help:       "sound loudness",
		Objectives: Objectives,
		MaxAge:     MaxAge,
	})
	soundReplayGain = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sound_replay_gain",
		Help:       "sound replay_gain",
		Objectives: Objectives,
		MaxAge:     MaxAge,
	})
	soundInstantPower = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sound_instant_power",
		Help:       "sound instant_power",
		Objectives: Objectives,
		MaxAge:     MaxAge,
	})
	soundRMS = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sound_rms",
		Help:       "sound rms",
		Objectives: Objectives,
		MaxAge:     MaxAge,
	})
	soundIntensity = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sound_intensity",
		Help:       "sound intensity",
		Objectives: Objectives,
		MaxAge:     MaxAge,
	})
)

func init() {
	prometheus.MustRegister(soundEnergy)
	prometheus.MustRegister(soundLoudness)
	prometheus.MustRegister(soundReplayGain)
	prometheus.MustRegister(soundInstantPower)
	prometheus.MustRegister(soundRMS)
	prometheus.MustRegister(soundIntensity)
}

func main() {
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)

	log.Print(run())
}

func run() error {
	err := portaudio.Initialize()
	if err != nil {
		return err
	}
	defer portaudio.Terminate()

	dev, err := portaudio.DefaultInputDevice()
	if err != nil {
		return err
	}

	frameSize := 4096
	framesc, err := ears.Listen(dev, frameSize, false)
	if err != nil {
		return err
	}

	anl := essentia.NewAnalyzer(frameSize)

	rec := &SM{
		process:    stateNone,
		threshold:  float32(*level),
		samplerate: int(dev.DefaultSampleRate),
	}

	for frames := range framesc {
		stats := anl.AnalyzeFrame(frames[0])

		soundEnergy.Observe(float64(stats.Energy))
		soundLoudness.Observe(float64(stats.Loudness))
		soundReplayGain.Observe(float64(stats.ReplayGain))
		soundInstantPower.Observe(float64(stats.InstantPower))
		soundRMS.Observe(float64(stats.RMS))
		soundIntensity.Observe(float64(stats.Intensity))

		logtrace(" %v", stats.Energy)
		rec.input(frames[0], stats.Energy)
	}

	return nil
}

func logtrace(patern string, args ...interface{}) {
	if *trace {
		log.Printf(patern, args...)
	}
}

/*************************** State Machine *********************************/
type Frame = []float32
type State int

const (
	none = iota
	increase
	active
	decrease
)

type SM struct {
	process    func(sm *SM, frame Frame, energy float32) State
	threshold  float32
	samplerate int
	frames     []Frame
	writer     *LameWriter
}

func (sm *SM) input(frame Frame, energy float32) {
	state := sm.process(sm, frame, energy)

	switch state {
	case none:
		logtrace("state none")
		sm.process = stateNone
	case increase:
		logtrace("state inc")
		sm.process = stateIncrease
	case active:
		logtrace("state write")
		sm.process = stateActive
	case decrease:
		logtrace("state dec")
		sm.process = stateDecrease

	}
}

func (sm *SM) free() State {
	if sm.writer != nil {
		sm.writer.Close()
		sm.writer = nil
	}
	return none
}

func stateNone(sm *SM, frame Frame, energy float32) State {
	if energy < sm.threshold {
		return none
	}

	sm.frames = []Frame{frame}
	return increase
}

func stateIncrease(sm *SM, frame Frame, energy float32) State {
	if energy < sm.threshold {
		return sm.free()
	}

	sm.frames = append(sm.frames, frame)
	if len(sm.frames) < 3 {
		return increase
	}

	// start mp3 file
	var err error
	sm.writer, err = NewLameWriter(sm.samplerate)
	if err != nil {
		// log error
		log.Printf("can't write mp3 file: %v", err)
		return sm.free()
	}

	for _, frame := range sm.frames {
		_, err := sm.writer.Write(frame)
		if err != nil {
			log.Printf("can't write mp3 file: %v", err)
			return sm.free()
		}
	}

	return active
}

func stateActive(sm *SM, frame Frame, energy float32) State {
	if energy < sm.threshold {
		sm.frames = []Frame{frame}
		return decrease
	}

	_, err := sm.writer.Write(frame)
	if err != nil {
		log.Printf("can't write mp3 file: %v", err)
		return sm.free()
	}

	return active
}

func stateDecrease(sm *SM, frame Frame, energy float32) State {
	if energy >= sm.threshold {
		for _, frame := range sm.frames {
			_, err := sm.writer.Write(frame)
			if err != nil {
				log.Printf("can't write mp3 file: %v", err)
				return sm.free()
			}
		}
		return active
	}

	sm.frames = append(sm.frames, frame)
	if len(sm.frames) > 20 {
		return sm.free()
	}

	return decrease
}

type LameWriter struct {
	out io.Writer
	enc *lame.Encoder
}

func NewLameWriter(samplerate int) (*LameWriter, error) {
	enc, err := lame.NewEncoder(samplerate) // create new or use static ?
	if err != nil {
		return nil, err
	}

	timestr := time.Now().Format(time.RFC3339)
	os.Mkdir("output", os.ModePerm)
	file, err := os.Create("output/" + timestr + ".mp3")
	if err != nil {
		return nil, err
	}

	return &LameWriter{file, enc}, nil
}

func (lw *LameWriter) Write(frame Frame) (int, error) {
	b, err := lw.enc.Encode(frame)
	if err != nil {
		return 0, err
	}

	_, err = lw.out.Write(b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (lw *LameWriter) Close() error {
	b := lw.enc.Flush()
	_, err := lw.out.Write(b)
	return err
}
