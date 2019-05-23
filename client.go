package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"fmt"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

// A Client represents a program connected to the PulseAudio daemon.
type Client struct {
	Index            int
	Name             string
	OwnerModuleIndex int
	Driver           string
	Properties       map[string]interface{}
	conn             *Conn
}

// Populate this client's fields with data in a string-interface{} map.
func (self *Client) Initialize(properties map[string]interface{}) error {
	self.Properties = properties
	return UnmarshalMap(self.Properties, self)
}

func (self *Client) P(key string) typeutil.Variant {
	return maputil.M(self.Properties).Get(key)
}

// Synchronize this client's data with the PulseAudio daemon.
func (self *Client) Refresh() error {
	operation := NewOperation(self.conn)
	defer operation.Destroy()

	operation.paOper = C.pa_context_get_client_info(
		self.conn.context,
		C.uint32_t(self.Index),
		(C.pa_client_info_cb_t)(C.pulse_get_client_info_callback),
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
			return fmt.Errorf("Invalid client response: expected 1 payload, got %d", l)
		}

		return nil

	})
}
