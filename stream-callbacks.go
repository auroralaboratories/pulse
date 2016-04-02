package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"errors"
	"fmt"
)

//export go_streamStateChange
func go_streamStateChange(streamId *C.char, message *C.char) {
	if obj, ok := cgoget(C.GoString(streamId)); ok {
		switch obj.(type) {
		case *Stream:
			stream := obj.(*Stream)

			if str := C.GoString(message); str == `` {
				stream.state <- nil
			} else {
				stream.state <- errors.New(str)
			}
		default:
			panic(fmt.Sprintf("go_streamStateChange(): invalid object %s; expected *pulse.Stream, got %T", streamId, obj))
		}
	}
}

//export go_streamPerformWrite
func go_streamPerformWrite(streamId *C.char, length C.size_t) {
	if obj, ok := cgoget(C.GoString(streamId)); ok {
		switch obj.(type) {
		case *Stream:
			stream := obj.(*Stream)

			stream.readFromSource(int(length))
		default:
			panic(fmt.Sprintf("go_streamPerformWrite(): invalid object %s; expected *pulse.Stream, got %T", streamId, obj))
		}
	}
}
