package darknet

// #include <stdlib.h>
// #include "class_name.h"
import "C"
import "unsafe"

func freeClassNames(names **C.char) {
	C.free_class_names(names)
}

func loadClassNames(dataConfigFile string) **C.char {
	d := C.CString(dataConfigFile)
	defer C.free(unsafe.Pointer(d))

	return C.read_class_names(d)
}

func makeClassNames(names **C.char, classes int) []string {
	out := make([]string, classes)
	for i := 0; i < classes; i++ {
		n := C.get_class_name(names, C.int(i), C.int(classes))
		s := C.GoString(n)
		out[i] = s
	}

	return out
}
