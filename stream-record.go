package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"io"
)

type RecordStream struct {
	*Stream
	io.Reader
}
