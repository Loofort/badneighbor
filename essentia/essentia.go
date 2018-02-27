package essentia

//// #cgo CXXFLAGS: -std=c++11

// #cgo LDFLAGS: -I/usr/local/include/essentia -I/usr/local/include/essentia/scheduler -I/usr/local/include/essentia/streaming -I/usr/local/include/essentia/utils -L/usr/local/lib -lessentia  -lfftw3 -lyaml -lavcodec -lavformat -lavutil -lavresample -lsamplerate -ltag -lfftw3f -lchromaprint
// #include "essentia.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

type cpointer struct {
	p       unsafe.Pointer
	destroy func(unsafe.Pointer)
}

func (cp *cpointer) free() {
	cp.destroy(cp.p)
	cp.p = nil
}

func newCPointer(p unsafe.Pointer, destroy func(unsafe.Pointer)) *cpointer {
	cp := &cpointer{p, destroy}
	runtime.SetFinalizer(cp, (*cpointer).free)
	return cp
}

func toGoArr(res C.struct_ResultArr) (unsafe.Pointer, int, error) {
	if res.Err != nil {
		str := C.GoString(res.Err)
		C.free(unsafe.Pointer(res.Err))

		err := errors.New(str)
		return nil, 0, err
	}

	if res.Cnt == 0 {
		return nil, 0, nil
	}

	return res.Res, int(res.Cnt), nil
}

/*************************************************************/

type Analyzer struct {
	*cpointer
}

func NewAnalyzer() Analyzer {

	ptr := C.NewAnalyzer()

	anl := Analyzer{newCPointer(ptr, destroyAnalyzer)}
	return anl
}
func destroyAnalyzer(p unsafe.Pointer) {
	C.DestroyAnalyzer(p)
}

func (anl Analyzer) AnalyzeFile(path string) ([]float32, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ptr, cnt, err := toGoArr(C.AnalyzeFile(anl.p, cpath))
	if err != nil {
		err = fmt.Errorf("can't analyze file: %v", err)
		return nil, err
	}
	if cnt == 0 {
		return nil, nil
	}
	defer C.free(ptr)

	slice := (*[1 << 30]float32)(ptr)[:cnt:cnt]
	result := make([]float32, cnt)
	copy(result, slice)

	return result, nil
}
