package main

import (
	"io"
	"os"

	mp3 "github.com/hajimehoshi/go-mp3"
)

func main() {

	ch := make(chan []byte)
	audioc := openAudioFile("testdata/one-two.mp3")
	go reader2chan(d, ch)

	statsc := getStatsAgregation(audioc)

	// each frame represent 10 ms of audio
	// so iteration should be faster
	for frame := range statsc {
		analyze()

		// save energies to prometheus
		save
	}

}

func openAudioFile(file string) (mp3.Decoder, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	return d, nil
}

func reader2chan(r io.Reader, ch chan []byte) {
	for {
		b := make([]byte, 1<<16)
		n, err := r.Read(b)
		if err != nil {
			close(ch)
			return
		}
		if n > 0 {
			ch <- b[0:n]
		}
	}
}
