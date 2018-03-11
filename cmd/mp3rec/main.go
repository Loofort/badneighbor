package main

import (
	"log"
	"os"
	"time"

	"github.com/Loofort/badneighbor/ears"
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

	outfile, err := os.Create("output.mp3")
	if err != nil {
		return err
	}

	enc, err := lame.NewEncoder(int(dev.DefaultSampleRate))
	if err != nil {
		return err
	}
	defer func() {
		b := enc.Flush()
		outfile.Write(b)
	}()

	tm := time.After(time.Second * 5)
	for {
		select {
		case <-tm:
			return nil
		case frames := <-framesc:
			b, err := enc.Encode(frames[0])
			if err != nil {
				return err
			}

			outfile.Write(b)
		}
	}

	return nil
}
