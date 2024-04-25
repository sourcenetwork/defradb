package connor

/*
#include <stdbool.h>
#include <stdlib.h>
#cgo LDFLAGS: -L./../libs -labi

typedef struct {
    char* data;
    size_t cap;
} OutputString;

extern bool match_conditions(const char* cond_json, const char* doc_json, OutputString* error);
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const strBufferSize = 256

func allocateOutputString() (C.OutputString, func()) {
	str := C.OutputString{
		data: (*C.char)(C.malloc(C.size_t(strBufferSize))),
		cap:  strBufferSize,
	}
	return str, func() { C.free(unsafe.Pointer(str.data)) }
}

func callMatchConditionsABI(conditions, data string) (bool, error) {
	cCond := C.CString(conditions)
	cDoc := C.CString(data)
	defer C.free(unsafe.Pointer(cCond))
	defer C.free(unsafe.Pointer(cDoc))

	var cError, freeCError = allocateOutputString()
	defer freeCError()

	result := C.match_conditions(cCond, cDoc, &cError)
	errorStr := C.GoString(cError.data)

	if errorStr == "" {
		return bool(result), nil
	}

	return false, fmt.Errorf(errorStr)
}
