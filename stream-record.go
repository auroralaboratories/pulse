package pulse

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
