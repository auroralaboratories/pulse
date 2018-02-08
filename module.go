package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"fmt"
	"unsafe"
)

// A Module represents PulseAudio drivers, configuration, and functionality
//
type Module struct {
	Argument string
	Client   *Client
	Index    uint
	Name     string

	properties map[string]interface{}
}

// Populate this module's fields with data in a string-interface{} map.
//
func (self *Module) Initialize(properties map[string]interface{}) error {
	self.properties = properties
	self.Index = uint(C.PA_INVALID_INDEX)

	if err := UnmarshalMap(self.properties, self); err == nil {
		return nil
	} else {
		return err
	}
}

// Synchronize this module's data with the PulseAudio daemon.
//
func (self *Module) Refresh() error {
	operation := NewOperation(self.Client)
	defer operation.Destroy()

	operation.paOper = C.pa_context_get_module_info(
		self.Client.context,
		C.uint32_t(self.Index),
		(C.pa_module_info_cb_t)(
			unsafe.Pointer(C.pulse_get_module_info_callback),
		),
		operation.Userdata(),
	)

	//  wait for the operation to finish and handle success and error cases
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

// Return whether the module is currently loaded or not
func (self *Module) IsLoaded() bool {
	return (self.Index != uint(C.PA_INVALID_INDEX))
}

// Load the module if it is not currently loaded
func (self *Module) Load() error {
	operation := NewOperation(self.Client)
	operation.paOper = C.pa_context_load_module(
		self.Client.context,
		C.CString(self.Name),
		C.CString(self.Argument),
		(C.pa_context_index_cb_t)(
			unsafe.Pointer(C.pulse_generic_index_callback),
		),
		operation.Userdata(),
	)

	//  wait for the operation to finish and handle success and error cases
	return operation.WaitSuccess(func(op *Operation) error {
		if err := UnmarshalMap(self.properties, self); err != nil {
			return err
		}

		return nil
	})
}

// Unload the module if it is currently loaded
func (self *Module) Unload() error {
	if self.IsLoaded() {
		operation := NewOperation(self.Client)
		operation.paOper = C.pa_context_unload_module(
			self.Client.context,
			C.uint32_t(self.Index),
			(C.pa_context_success_cb_t)(
				unsafe.Pointer(C.pulse_generic_success_callback),
			),
			operation.Userdata(),
		)

		//  wait for the operation to finish and handle success and error cases
		return operation.WaitSuccess(func(op *Operation) error {
			self.Index = uint(C.PA_INVALID_INDEX)
			return nil
		})
	} else {
		return fmt.Errorf("The '%s' module is already unloaded", self.Name)
	}
}
