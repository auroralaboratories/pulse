package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"io"
)

type UploadStream struct {
	*Stream
	io.Writer
}
