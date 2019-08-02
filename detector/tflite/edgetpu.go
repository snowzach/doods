package tflite

/*
#define _GNU_SOURCE
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <tensorflow/lite/experimental/c/c_api.h>
#include <tensorflow/lite/experimental/c/c_api_experimental.h>
#include <tensorflow/lite/experimental/c/c_api_types.h>
#cgo windows CFLAGS: -D__LITTLE_ENDIAN__
#cgo CFLAGS: -I/opt/tensorflow
#cgo windows CXXFLAGS: -D__LITTLE_ENDIAN__
#cgo CXXFLAGS: -std=c++11
#cgo CXXFLAGS: -I/opt/tensorflow
#cgo LDFLAGS: -ltensorflowlite_c -ledgetpu
#cgo linux LDFLAGS: -ltensorflowlite_c -ltensorflowlite -ledgetpu

int HasEdgeTPU();
TfLiteRegistration* RegisterEdgeTPUCustomOp();
void EdgeTPUSetup(TFL_Interpreter *i, TFL_Model *m);
*/
import "C"
import (
	"reflect"
	"unsafe"
)

var EdgeTPUCustomOp = "edgetpu-custom-op"

func wrap(p *C.TfLiteRegistration) *ExpRegistration {
	return &ExpRegistration{
		Init:            unsafe.Pointer(reflect.ValueOf(p.init).Pointer()),
		Free:            unsafe.Pointer(reflect.ValueOf(p.free).Pointer()),
		Prepare:         unsafe.Pointer(reflect.ValueOf(p.prepare).Pointer()),
		Invoke:          unsafe.Pointer(reflect.ValueOf(p.invoke).Pointer()),
		ProfilingString: unsafe.Pointer(reflect.ValueOf(p.profiling_string).Pointer()),
	}
}

// HasEdgeTPU returns true if an edgetpu was detected
func HasEdgeTPU() bool {
	return C.HasEdgeTPU() != 0
}

// Register_EdgeTPU custom op registration function for the edge TPU
func Register_EdgeTPU() *ExpRegistration {
	return wrap(C.RegisterEdgeTPUCustomOp())
}

// This configures the interpreter and model for the edgetpu
func EdgeTPUSetup(i *Interpreter, m *Model) {
	C.EdgeTPUSetup(i.i, m.m)
}
