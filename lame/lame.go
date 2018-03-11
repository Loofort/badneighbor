package lame

import (
	"fmt"
	"runtime"
	"unsafe"
)

/*
#cgo LDFLAGS: -lmp3lame
#include "lame/lame.h"
*/
import "C"

type Encoder struct {
	Lame *C.struct_lame_global_struct
}

func NewEncoder(sampleRate int) (*Encoder, error) {
	lm := C.lame_init()
	enc := &Encoder{lm}
	runtime.SetFinalizer(enc, finalize)

	C.lame_set_in_samplerate(lm, C.int(sampleRate))
	C.lame_set_VBR(lm, C.vbr_default)
	C.lame_set_VBR_quality(lm, 3)
	if ret := C.lame_init_params(lm); ret < 0 {
		return nil, fmt.Errorf("Error occurred during Lame initializing. Code = %d", ret)
	}

	return enc, nil
}

func finalize(enc *Encoder) {
	C.lame_close(enc.Lame)
}

func (enc *Encoder) Encode(samples []float32) ([]byte, error) {
	// inSample * (inRate / outRate) / (inNumChan / outNumChan)
	//InNumChannels := 1
	//outNumChannels := 1 // mono
	//numSamples := int(int64(len(samples)) * int64(enc.InSampleRate) / int64(44100) * int64(outNumChannels) / int64(InNumChannels))
	mp3BufSize := int(1.25*float32(len(samples)*4) + 7200 + 1) // follow the instruction from LAME
	mp3Buf := make([]byte, mp3BufSize)

	cIn := (*C.float)(unsafe.Pointer(&samples[0]))
	cOut := (*C.uchar)(unsafe.Pointer(&mp3Buf[0]))
	ret := C.lame_encode_buffer_ieee_float(
		enc.Lame,
		cIn,
		cIn,
		C.int(len(samples)),
		cOut,
		C.int(len(mp3Buf)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("Error occurred during Lame encoding. Code = %d", ret)
	}

	return mp3Buf[:int(ret)], nil
}

func (enc *Encoder) Flush() []byte {
	out := make([]byte, 7200) // the buffer size which is safe to hold possible data at a time
	cOut := (*C.uchar)(unsafe.Pointer(&out[0]))
	ret := C.lame_encode_flush(
		enc.Lame,
		cOut,
		C.int(len(out)),
	)

	return out[0:int(ret)]
}
