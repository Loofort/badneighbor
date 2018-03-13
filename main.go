package main

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/Loofort/badneighbor/ears"
	"github.com/Loofort/badneighbor/essentia"
	"github.com/Loofort/badneighbor/lame"
	"github.com/gordonklaus/portaudio"
)

const (
	listen = iota
	record
)

func main() {
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
		threshold:  0.1,
		samplerate: int(dev.DefaultSampleRate),
	}

	for frames := range framesc {
		energy := anl.FrameEnergy(frames[0])
		rec.input(frames[0], energy)
	}

	return nil
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
		sm.process = stateNone
	case increase:
		sm.process = stateIncrease
	case active:
		sm.process = stateActive
	case decrease:
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
