package pulse

// #cgo CFLAGS: -Wno-implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"errors"
)

//export go_streamStateChange
func go_streamStateChange(streamId *C.char, message *C.char) {
	v := cgoget(C.GoString(streamId))

	if stream, ok := v.(*Stream); ok {
		if str := C.GoString(message); str == `` {
			stream.state <- nil
		} else {
			stream.state <- errors.New(str)
		}
	}
}

//export go_streamPerformWrite
func go_streamPerformWrite(streamId *C.char, length C.size_t) {
	if stream, ok := cgoget(C.GoString(streamId)).(*Stream); ok {
		stream.readFromSource(int(length))
	}
}
