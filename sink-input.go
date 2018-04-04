package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "client.h"
// #cgo pkg-config: libpulse
import "C"
import "fmt"

// A SinkInput represents client ends of streams inside the server, i.e. they
// connect a client stream to one of the global sinks.
//
type SinkInput struct {
	Client      *Client
	Index       int
	Name        string
	SinkIndex   int
	ModuleIndex int
	ClientIndex int
	Muted       bool
	properties  map[string]interface{}
}

// Populate this sink inputs's fields with data in a string-interface{} map.
//
func (self *SinkInput) Initialize(properties map[string]interface{}) error {
	self.properties = properties

	if err := UnmarshalMap(self.properties, self); err != nil {
		return err
	}

	return nil
}

// Synchronize this sink input's data with the PulseAudio daemon.
//
func (self *SinkInput) Refresh() error {
	operation := NewOperation(self.Client)
	defer operation.Destroy()

	operation.paOper = C.pa_context_get_sink_input_info(
		self.Client.context,
		C.uint32_t(self.Index),
		(C.pa_sink_info_cb_t)(C.pulse_get_sink_input_info_by_index_callback),
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
			return fmt.Errorf("Invalid sink response: expected 1 payload, got %d", l)
		}

		return nil

	})
}

func (self *SinkInput) MoveToSink(sinkIndex int) error {
	operation := NewOperation(self.Client)
	defer operation.Destroy()

	//  make the call
	operation.paOper = C.pa_context_move_sink_input_by_index(
		self.Client.context,
		C.uint32_t(self.Index),
		C.uint32_t(sinkIndex),
		(C.pa_context_success_cb_t)(C.pulse_generic_success_callback),
		operation.Userdata(),
	)

	//  wait for the result, refresh, return any errors
	if err := operation.Wait(); err == nil {
		return self.Refresh()
	} else {
		return err
	}
}
