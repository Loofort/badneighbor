package essentia

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestAFile(t *testing.T) {

	anl := NewAnalyzer()
	res, err := anl.AnalyzeFile("../testdata/one-two.mp3")
	if err != nil {
		t.Fatal(err)
	}

	//outmetrics(res)
	outtext(res)
}
func outtext(res []float32) {
	epoch := time.Now().Add(time.Duration(len(res)) * 10 * time.Millisecond * -1)
	for i, v := range res {
		fmt.Printf("sound_energy %f %d\n", v, makeTimestamp(epoch, i))
	}
}

func makeTimestamp(epoch time.Time, i int) int64 {
	now := epoch.Add(time.Duration(i) * 10 * time.Millisecond)
	return now.UnixNano() / int64(time.Millisecond)
}

func outmetrics(res []float32) {
	soundEnergy := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sound_energy",
		Help: "instant energy",
	})
	prometheus.MustRegister(soundEnergy)

	//fmt.Printf("GORES: %v\n", res)
	for _, r := range res {
		soundEnergy.Set(float64(r))
	}

	enc := expfmt.NewEncoder(os.Stdout, expfmt.FmtText)

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		//http.Error(w, "An error has occurred during metrics collection:\n\n"+err.Error(), http.StatusInternalServerError)
		//return
		panic("aaa")
	}

	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			//http.Error(w, "An error has occurred during metrics encoding:\n\n"+err.Error(), http.StatusInternalServerError)
			//return
			panic("bbb")
		}
	}

}
