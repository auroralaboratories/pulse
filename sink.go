package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "unsafe"
)

type SinkState int
const (
    SinkStateInvalid   SinkState = C.PA_SINK_INVALID_STATE
    SinkStateRunning             = C.PA_SINK_RUNNING
    SinkStateIdle                = C.PA_SINK_IDLE
    SinkStateSuspended           = C.PA_SINK_SUSPENDED
)

// A Sink represents a logical audio output destination with its own volume control.
//
type Sink struct {
    Client             *Client
    CardIndex          int
    Description        string
    DriverName         string
    Index              int
    ModuleIndex        int
    MonitorSourceIndex int
    MonitorSourceName  string
    Muted              bool
    Name               string
    NumFormats         int
    NumPorts           int
    NumVolumeSteps     int
    State              SinkState
    Channels           int
    CurrentVolumeStep  int
    VolumeFactor       float64

    properties         map[string]interface{}
}


// Populate this sink's fields with data in a string-interface{} map.
//
func (self *Sink) Initialize(properties map[string]interface{}) error {
    self.properties = properties

    if err := UnmarshalMap(self.properties, self); err == nil {
        self.loadSinkStateFromProperties()
    }else{
        return err
    }

    return nil
}

func (self *Sink) loadSinkStateFromProperties() {
    state := SinkStateInvalid

    if v, ok := self.properties[`_state`]; ok {
        switch v.(type) {
        case int64:
            switch int(v.(int64)) {
            case int(C.PA_SINK_RUNNING):
                state = SinkStateRunning
            case int(C.PA_SINK_IDLE):
                state = SinkStateIdle
            case int(C.PA_SINK_SUSPENDED):
                state = SinkStateSuspended
            }
        }
    }

    self.State = state
}


// Synchronize this sink's data with the PulseAudio daemon.
//
func (self *Sink) Refresh() error {
    operation := NewOperation(self.Client)

    operation.paOper = C.pa_context_get_sink_info_by_index(C.pulse_get_context(), C.uint32_t(self.Index), (C.pa_sink_info_cb_t)(unsafe.Pointer(C.pulse_get_sink_info_by_index_callback)), unsafe.Pointer(operation))

//  wait for the operation to finish and handle success and error cases
    return operation.WaitSuccess(func(op *Operation) error {
        if l := len(op.Payloads); l == 1 {
            payload := operation.Payloads[0]

            if err := self.Initialize(payload.Properties); err != nil {
                return err
            }
        }else{
            return fmt.Errorf("Invalid sink response: expected 1 payload, got %d", l)
        }

        return nil

    })
}


// Set the volume of all channels of this sink to a factor of the maximum
// volume (0.0 <= v <= 1.0).  Factors greater than 1.0 will be accepted, but
// clipping or distortion may occur beyond that value.
//
func (self *Sink) SetVolume(factor float64) error {
    if self.Channels > 0 {
        operation  := NewOperation(self.Client)
        newVolume  := &C.pa_cvolume{}

    //  new volume is the (maximum number of normal volume steps * factor)
        newVolume   = C.pa_cvolume_init(newVolume)
        newVolumeT := C.pa_volume_t(C.uint32_t(uint(float64(self.NumVolumeSteps) * factor)))

    //  prepare newVolume for its journey into PulseAudio
        C.pa_cvolume_set(newVolume, C.uint(self.Channels), newVolumeT)

    //  make the call
        operation.paOper = C.pa_context_set_sink_volume_by_index(C.pulse_get_context(), C.uint32_t(self.Index), newVolume, (C.pa_context_success_cb_t)(unsafe.Pointer(C.pulse_generic_success_callback)), unsafe.Pointer(operation))

    //  wait for the result, refresh, return any errors
        return operation.WaitSuccess(func(op *Operation) error {
            return self.Refresh()
        })
    }else{
        return fmt.Errorf("Cannot set volume on sink %d, no channels defined", self.Index)
    }

    return nil
}


// Add the given factor to the current sink volume
//
func (self *Sink) IncreaseVolume(factor float64) error {
    if err := self.Refresh(); err == nil {
        newFactor := (self.VolumeFactor + factor)
        return self.SetVolume(newFactor)
    }else{
        return err
    }
}


// Remove the given factor from the current sink volume, or
// set to a minimum of 0.0.
//
func (self *Sink) DecreaseVolume(factor float64) error {
    if err := self.Refresh(); err == nil {
        newFactor := (self.VolumeFactor - factor)

        if newFactor < 0.0 {
            return self.SetVolume(0.0)
        }else{
            return self.SetVolume(newFactor)
        }
    }else{
        return err
    }
}


//  Explicitly set the muted or unmuted state of the sink.
//
func (self *Sink) SetMute(mute bool) error {
    operation  := NewOperation(self.Client)

    var muting C.int

    if mute {
        muting = C.int(1)
    }else{
        muting = C.int(0)
    }

    operation.paOper = C.pa_context_set_sink_mute_by_index(C.pulse_get_context(), C.uint32_t(self.Index), muting, (C.pa_context_success_cb_t)(unsafe.Pointer(C.pulse_generic_success_callback)), unsafe.Pointer(operation))

//  wait for the result, refresh, return any errors
    return operation.WaitSuccess(func(op *Operation) error {
        return self.Refresh()
    })
}


// Explicitly mute the sink.
//
func (self *Sink) Mute() error {
    return self.SetMute(true)
}


// Explicitly unmute the sink.
//
func (self *Sink) Unmute() error {
    return self.SetMute(false)
}


// Mute or unmute the sink, depending on whether it is currently
// unmuted or muted (respectively).
//
func (self *Sink) ToggleMute() error {
    if err := self.Refresh(); err == nil {
        return self.SetMute(!self.Muted)
    }else{
        return err
    }
}