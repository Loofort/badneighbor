package essentia

//// #cgo CXXFLAGS: -std=c++11

// #cgo LDFLAGS: -I/usr/local/include/essentia -I/usr/local/include/essentia/scheduler -I/usr/local/include/essentia/streaming -I/usr/local/include/essentia/utils -L/usr/local/lib -lessentia  -lfftw3 -lyaml -lavcodec -lavformat -lavutil -lavresample -lsamplerate -ltag -lfftw3f -lchromaprint
// #include "essentia.h"
// #include <stdlib.h>
import "C"
import (
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

func (anl Analyzer) AnalyzeFile(path string) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	C.AnalyzeFile(anl.p, cpath)
}
