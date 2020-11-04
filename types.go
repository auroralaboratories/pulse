package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

const DefaultVolumeStep = 65536

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

type Volume struct {
	Name  string
	Value float64
}
