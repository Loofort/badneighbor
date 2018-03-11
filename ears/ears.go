package ears

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/gordonklaus/portaudio"
)

var tmpl = template.Must(template.New("").Parse(
	`{{. | len}} host APIs: {{range .}}{{if .DefaultInputDevice}}
	Name:                   {{.Name}}
	Default input device:   {{.DefaultInputDevice.Name}}
	Devices: {{range .Devices}}{{if gt .MaxInputChannels 0}}
		Name:                      {{.Name}}
		MaxInputChannels:          {{.MaxInputChannels}}
		DefaultLowInputLatency:    {{.DefaultLowInputLatency}}
		DefaultHighInputLatency:   {{.DefaultHighInputLatency}}
		DefaultSampleRate:         {{.DefaultSampleRate}}
	{{end}}{{end}}
{{end}}{{end}}`,
))

func PrintInputDevices() error {
	err := portaudio.Initialize()
	if err != nil {
		return err
	}
	defer portaudio.Terminate()
	hs, err := portaudio.HostApis()
	if err != nil {
		return err
	}

	err = tmpl.Execute(os.Stdout, hs)
	return err
}

func Listen(dev *portaudio.DeviceInfo, frameSize int, stereo bool) (chan [][]float32, error) {
	sdp := portaudio.StreamDeviceParameters{
		Device:   dev,
		Latency:  dev.DefaultHighInputLatency,
		Channels: 1,
	}
	if stereo {
		sdp.Channels = 2
	}
	if sdp.Channels > dev.MaxInputChannels {
		return nil, fmt.Errorf("can open sound channels, for %v MaxInputChannels is %d", dev.Name, dev.MaxInputChannels)
	}

	sp := portaudio.StreamParameters{
		Input:           sdp,
		SampleRate:      dev.DefaultSampleRate,
		FramesPerBuffer: frameSize, // portaudio.FramesPerBufferUnspecified (0) - special value for better performance
	}

	buf := [][]float32{}
	buf = append(buf, make([]float32, frameSize))
	if stereo {
		buf = append(buf, make([]float32, frameSize))
	}

	stream, err := portaudio.OpenStream(sp, buf)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	// read stream synchronously to obtain defined frame length
	// note: we never know if sound is throttled (we need to use callback to deal with it)
	samplesc := make(chan [][]float32, 1)
	go readLoop(stream, buf, samplesc)

	return samplesc, nil
}

func readLoop(stream *portaudio.Stream, buf [][]float32, samplesc chan [][]float32) {
	defer stream.Stop()

	for {
		err := stream.Read()
		if err != nil {
			// todo: write metric or log
			log.Printf("can't open audio input stream: %v\n", err)
			continue
		}

		bufcopy := [][]float32{}
		for _, b := range buf {
			bc := make([]float32, len(b))
			if n := copy(bc, b); n == 0 {
				continue
			}
			bufcopy = append(bufcopy, bc)
		}

		select {
		case samplesc <- bufcopy:
		default:
			// todo: write metric or log
			//log.Printf("throttling of input samples is detected\n")
			fmt.Printf("[T]")
		}
	}
}
