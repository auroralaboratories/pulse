package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"time"

	"github.com/ghetzel/go-stockutil/maputil"
)

func populateStruct(data map[string]interface{}, target interface{}) error {
	return maputil.TaggedStructFromMap(data, target, ``)
}

func MaxDuration() time.Duration {
	return time.Duration(int64((1 << 63) - 1))
}
