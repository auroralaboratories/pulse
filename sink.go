// +build !cgocheck
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

type SinkState int
const (
    SinkStateInvalid   SinkState = C.PA_SINK_INVALID_STATE
    SinkStateRunning             = C.PA_SINK_RUNNING
    SinkStateIdle                = C.PA_SINK_IDLE
    SinkStateSuspended           = C.PA_SINK_SUSPENDED
)

type Sink struct {
    Client             *Client
    CardIndex          int
    Description        string
    DriverName         string
    Index              int
    ModuleIndex        int
    MonitorSourceIndex int
    MonitorSourceName  string
    Mute               bool
    Name               string
    NumFormats         int
    NumPorts           int
    NumVolumeSteps     int
    State              SinkState
}

