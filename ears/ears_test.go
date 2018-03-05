package ears

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gordonklaus/portaudio"
)

func TestPrintDevices(t *testing.T) {
	err := PrintInputDevices()
	require.NoError(t, err)
}
func TestInputStream(t *testing.T) {
	err := portaudio.Initialize()
	require.NoError(t, err)
	defer portaudio.Terminate()

	dev, err := portaudio.DefaultInputDevice()
	require.NoError(t, err)

	samplesc, err := Listen(dev, 1024, true)
	require.NoError(t, err)

	cnt := 0
	for buf := range samplesc {
		for i, samples := range buf {
			//_ := binary.Write(f, binary.BigEndian, in)
			fmt.Printf("chan %d: %v \n", i, samples)
		}
		fmt.Printf("\n")
		if cnt++; cnt > 3 {
			break
		}
	}
}
