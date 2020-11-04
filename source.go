package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"encoding/json"
	"fmt"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type SourceState int

const (
	SourceStateInvalid   SourceState = C.PA_SOURCE_INVALID_STATE
	SourceStateRunning               = C.PA_SOURCE_RUNNING
	SourceStateIdle                  = C.PA_SOURCE_IDLE
	SourceStateSuspended             = C.PA_SOURCE_SUSPENDED
)

func (self SourceState) String() string {
	switch self {
	case SourceStateRunning:
		return `RUNNING`
	case SourceStateIdle:
		return `IDLE`
	case SourceStateSuspended:
		return `SUSPENDED`
	default:
		return `INVALID`
	}
}

// A Source represents a logical audio input source
//
type Source struct {
	BaseVolumeStep     int
	CardIndex          int
	Channels           int
	CurrentVolumeStep  int
	Description        string
	DriverName         string
	Index              int
	ModuleIndex        int
	MonitorOfSinkIndex int
	MonitorOfSinkName  string
	Muted              bool
	Name               string
	NumFormats         int
	NumPorts           int
	NumVolumeSteps     int
	State              SourceState
	VolumeFactor       float64
	Properties         map[string]interface{}
	conn               *Conn
}

func (self *Source) MarshalJSON() ([]byte, error) {
	type Alias Source

	return json.Marshal(&struct {
		StateValue string
		*Alias
	}{
		StateValue: self.State.String(),
		Alias:      (*Alias)(self),
	})
}

// Populate this source's fields with data in a string-interface{} map.
//
func (self *Source) Initialize(properties map[string]interface{}) error {
	self.Properties, _ = maputil.DiffuseMap(properties, `.`)

	if err := populateStruct(self.Properties, self); err == nil {
		self.loadSourceStateFromProperties()
	} else {
		return err
	}

	return nil
}

func (self *Source) P(key string) typeutil.Variant {
	return maputil.M(self.Properties).Get(key)
}

func (self *Source) loadSourceStateFromProperties() {
	state := SourceStateInvalid

	if v := self.P(`_state`); !v.IsNil() {
		switch int(v.Int()) {
		case int(C.PA_SOURCE_RUNNING):
			state = SourceStateRunning
		case int(C.PA_SOURCE_IDLE):
			state = SourceStateIdle
		case int(C.PA_SOURCE_SUSPENDED):
			state = SourceStateSuspended
		}

		delete(self.Properties, `_state`)
	}

	self.State = state
}

// Synchronize this source's data with the PulseAudio daemon.
//
func (self *Source) Refresh() error {
	operation := NewOperation(self.conn)
	defer operation.Destroy()

	operation.paOper = C.pa_context_get_source_info_by_index(self.conn.context, C.uint32_t(self.Index), (C.pa_source_info_cb_t)(C.pulse_get_source_info_by_index_callback), operation.Userdata())

	// wait for the operation to finish and handle success and error cases
	return operation.WaitSuccess(func(op *Operation) error {
		if l := len(op.Payloads); l == 1 {
			payload := operation.Payloads[0]

			if err := self.Initialize(payload.Properties); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Invalid source response: expected 1 payload, got %d", l)
		}

		return nil

	})
}

// Set the volume of all channels of this source to a factor of the maximum
// volume (0.0 <= v <= 1.0).  Factors greater than 1.0 will be accepted, but
// clipping or distortion may occur beyond that value.
//
func (self *Source) SetVolume(factor float64) error {
	if self.Channels > 0 {
		operation := NewOperation(self.conn)
		defer operation.Destroy()
		newVolume := &C.pa_cvolume{}

		// new volume is the (maximum number of normal volume steps * factor)
		newVolume = C.pa_cvolume_init(newVolume)
		newVolumeT := C.pa_volume_t(C.uint32_t(uint(float64(self.NumVolumeSteps) * factor)))

		// prepare newVolume for its journey into PulseAudio
		C.pa_cvolume_set(newVolume, C.uint(self.Channels), newVolumeT)

		// make the call
		operation.paOper = C.pa_context_set_source_volume_by_index(self.conn.context, C.uint32_t(self.Index), newVolume, (C.pa_context_success_cb_t)(C.pulse_generic_success_callback), operation.Userdata())

		// wait for the result, refresh, return any errors
		if err := operation.Wait(); err == nil {
			return self.Refresh()
		} else {
			return err
		}
	} else {
		return fmt.Errorf("Cannot set volume on source %d, no channels defined", self.Index)
	}
}

// Add the given factor to the current source volume
//
func (self *Source) IncreaseVolume(factor float64) error {
	if err := self.Refresh(); err == nil {
		newFactor := (self.VolumeFactor + factor)
		return self.SetVolume(newFactor)
	} else {
		return err
	}
}

// Remove the given factor from the current source volume, or
// set to a minimum of 0.0.
//
func (self *Source) DecreaseVolume(factor float64) error {
	if err := self.Refresh(); err == nil {
		newFactor := (self.VolumeFactor - factor)

		if newFactor < 0.0 {
			return self.SetVolume(0.0)
		} else {
			return self.SetVolume(newFactor)
		}
	} else {
		return err
	}
}

// Explicitly set the muted or unmuted state of the source.
//
func (self *Source) SetMute(mute bool) error {
	operation := NewOperation(self.conn)
	defer operation.Destroy()

	var muting C.int

	if mute {
		muting = C.int(1)
	} else {
		muting = C.int(0)
	}

	operation.paOper = C.pa_context_set_source_mute_by_index(self.conn.context, C.uint32_t(self.Index), muting, (C.pa_context_success_cb_t)(C.pulse_generic_success_callback), operation.Userdata())

	// wait for the result, refresh, return any errors
	if err := operation.Wait(); err == nil {
		return self.Refresh()
	} else {
		return err
	}
}

// Explicitly mute the source.
//
func (self *Source) Mute() error {
	return self.SetMute(true)
}

// Explicitly unmute the source.
//
func (self *Source) Unmute() error {
	return self.SetMute(false)
}

// Mute or unmute the source, depending on whether it is currently
// unmuted or muted (respectively).
//
func (self *Source) ToggleMute() error {
	if err := self.Refresh(); err == nil {
		return self.SetMute(!self.Muted)
	} else {
		return err
	}
}
