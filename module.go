package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

var NoSuchModuleErr = fmt.Errorf("no such module")

func IsNoSuchModuleErr(err error) bool {
	if err != nil && err == NoSuchModuleErr {
		return true
	}

	return false
}

// A Module represents PulseAudio drivers, configuration, and functionality
//
type Module struct {
	Argument   string
	Index      uint
	Name       string
	Properties map[string]interface{}
	conn       *Conn
}

// Populate this module's fields with data in a string-interface{} map.
func (self *Module) Initialize(properties map[string]interface{}) error {
	self.Properties, _ = maputil.DiffuseMap(properties, `.`)
	self.Index = uint(C.PA_INVALID_INDEX)

	return populateStruct(self.Properties, self)
}

func (self *Module) P(key string) typeutil.Variant {
	return maputil.M(self.Properties).Get(key)
}

// Synchronize this module's data with the PulseAudio daemon.
func (self *Module) Refresh() error {
	operation := NewOperation(self.conn)
	defer operation.Destroy()

	operation.paOper = C.pa_context_get_module_info(
		self.conn.context,
		C.uint32_t(self.Index),
		(C.pa_module_info_cb_t)(unsafe.Pointer(C.pulse_get_module_info_callback)),
		operation.Userdata(),
	)

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

// Return whether the module is currently loaded or not
func (self *Module) IsLoaded() bool {
	return (self.Index != uint(C.PA_INVALID_INDEX))
}

// Load the module if it is not currently loaded
func (self *Module) Load() error {
	operation := NewOperation(self.conn)
	operation.paOper = C.pa_context_load_module(
		self.conn.context,
		C.CString(self.Name),
		C.CString(self.Argument),
		(C.pa_context_index_cb_t)(unsafe.Pointer(C.pulse_generic_index_callback)),
		operation.Userdata(),
	)

	// wait for the operation to finish and handle success and error cases
	return operation.WaitSuccess(func(op *Operation) error {
		if err := populateStruct(self.Properties, self); err != nil {
			return err
		}

		return nil
	})
}

// Unload the module if it is currently loaded
func (self *Module) Unload() error {
	if self.IsLoaded() {
		operation := NewOperation(self.conn)
		operation.paOper = C.pa_context_unload_module(
			self.conn.context,
			C.uint32_t(self.Index),
			(C.pa_context_success_cb_t)(unsafe.Pointer(C.pulse_generic_success_callback)),
			operation.Userdata(),
		)

		// wait for the operation to finish and handle success and error cases
		return operation.WaitSuccess(func(op *Operation) error {
			self.Index = uint(C.PA_INVALID_INDEX)
			return nil
		})
	} else {
		return fmt.Errorf("The '%s' module is already unloaded", self.Name)
	}
}
