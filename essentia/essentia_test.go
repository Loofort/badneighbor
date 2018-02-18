package essentia

import "testing"

func TestAFile(t *testing.T) {

	anl := NewAnalyzer()
	anl.AnalyzeFile("../testdata/one-two.mp3")
}
