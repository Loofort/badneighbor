package main

import (
	"fmt"
	"log"

	"github.com/Loofort/badneighbor/ears"
	"github.com/Loofort/badneighbor/essentia"
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
	//state := listen
	res := []float32{}
	cnt := 500
	for frames := range framesc {
		cnt--
		if cnt == 0 {
			break
		}

		energy := anl.FrameEnergy(frames[0])
		res = append(res, energy)

	}

	fmt.Printf("res %v\n", res)
	return nil
}

type Frame = []float32
type Event = int

const(
	none = iota
	increase
	active
	deccrease
)
type SM {
   state int
   threshold float32
   buf Frame
   writer *lame.LameWriter

   state func(Frame) Event
   transition map[int]int
}

func (sm *SM) input(frame Frame) {
	energy := anl.FrameEnergy(frames[0])
	event := sm.state(frame, energy)
	sm.state := sm.transition[sm.state][event]
}

func (sm *SM) stateNone(frame Frame, energy float32) Event {
	if energy < sm.threshold {
		return none
	}

	sm.frames = append(sm.frames, frame)
	return increase
}

func (sm *SM) stateIncrease(frame Frame, energy float32) Event {
	if energy < sm.threshold {
		sm.frames = []Frame{}
		return none
	}

	sm.frames = append(sm.frames, frame)
	if len(sm.frames) < 3 {
		return increase
	}

	for _, frame := range sm.frames {
		int, err := sm.writer.Write(p []byte) 
	}
	return active
}

