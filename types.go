package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"
import (
	"github.com/ghetzel/go-stockutil/typeutil"
)

type ServerInfo struct {
	Channels               int
	Cookie                 int
	DaemonHostname         string
	DaemonUser             string
	DefaultSinkName        string
	DefaultSourceName      string
	LibraryProtocolVersion int
	Name                   string
	ProtocolVersion        int
	SampleFormat           string
	SampleRate             int
	ServerString           string
	Version                string
}

type Volume float64

func (self Volume) Convert(in interface{}) (interface{}, error) {
	return (typeutil.Float(in) / 65536) * 100.0, nil
}

type VolumeSet map[string]interface{}

func (self VolumeSet) Convert(in interface{}) (interface{}, error) {
	self = make(VolumeSet)

	for k, v := range typeutil.MapNative(in) {
		self[k] = (typeutil.Float(v) / 65536) * 100.0
	}

	return self, nil
}
