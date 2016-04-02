// Golang bindings for PulseAudio 8.x+
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"time"
	"unsafe"
)

type ClientLockFunc func() error

type ContextState int

const (
	StateUnconnected ContextState = C.PA_CONTEXT_UNCONNECTED  // The context hasn't been connected yet.
	StateConnecting               = C.PA_CONTEXT_CONNECTING   // A connection is being established.
	StateAuthorizing              = C.PA_CONTEXT_AUTHORIZING  // The client is authorizing itself to the daemon.
	StateSettingName              = C.PA_CONTEXT_SETTING_NAME // The client is passing its application name to the daemon.
	StateReady                    = C.PA_CONTEXT_READY        // The connection is established, the context is ready to execute operations.
	StateFailed                   = C.PA_CONTEXT_FAILED       // The connection failed or was disconnected.
	StateTerminated               = C.PA_CONTEXT_TERMINATED   // The connection was terminated cleanly.
)

// A PulseAudio Client represents a connection to a PulseAudio daemon (either locally or
// on a remote host). A Client is the primary entry point for working with PulseAudio
// objects and data.
//
type Client struct {
	ID               string
	Name             string
	Server           string
	OperationTimeout time.Duration

	state    chan error
	mainloop *C.pa_threaded_mainloop
	context  *C.pa_context
	api      *C.pa_mainloop_api
}

func NewClient(name string) (*Client, error) {
	rv := &Client{
		ID:               uuid.NewV4().String(),
		Name:             name,
		OperationTimeout: (time.Duration(DEFAULT_OPERATION_TIMEOUT_MSEC) * time.Millisecond),

		state: make(chan error),
	}

	cgoregister(rv.ID, rv)

	rv.mainloop = C.pa_threaded_mainloop_new()
	if rv.mainloop == nil {
		return nil, fmt.Errorf("Failed to create PulseAudio mainloop")
	}

	rv.api = C.pa_threaded_mainloop_get_api(rv.mainloop)
	rv.context = C.pa_context_new(rv.api, C.CString(name))

	C.pa_context_set_state_callback(rv.context, (C.pa_context_notify_cb_t)(C.pulse_context_state_callback), rv.ToUserdata())

	//  lock the mainloop until the context is ready
	rv.Lock()

	//  start the mainloop
	rv.Start()

	//  initiate context connect
	if int(C.pa_context_connect(rv.context, nil, (C.pa_context_flags_t)(0), nil)) != 0 {
		defer rv.Stop()
		defer rv.Destroy()

		return nil, rv.GetLastError()
	}

	//  wait for context to be ready
	for {
		state := ContextState(int(C.pa_context_get_state(rv.context)))
		breakOut := false

		switch state {
		case StateUnconnected, StateConnecting, StateAuthorizing, StateSettingName:
			if err := rv.Wait(); err != nil {
				return nil, err
			}
		case StateFailed:
			return nil, rv.GetLastError()
		case StateTerminated:
			return nil, fmt.Errorf("PulseAudio connection was terminated during setup")
		case StateReady:
			breakOut = true
		default:
			return nil, fmt.Errorf("Encountered unknown connection state %d during setup", state)
		}

		if breakOut {
			break
		}
	}

	rv.Unlock()

	return rv, nil
}

// Retrieve information about the connected PulseAudio daemon
//
func (self *Client) GetServerInfo() (ServerInfo, error) {
	operation := NewOperation(self)
	defer operation.Destroy()

	info := ServerInfo{}

	operation.paOper = C.pa_context_get_server_info(self.context, (C.pa_server_info_cb_t)(unsafe.Pointer(C.pulse_get_server_info_callback)), operation.ToUserdata())

	//  wait for the operation to finish and handle success and error cases
	return info, operation.WaitSuccess(func(op *Operation) error {
		if len(op.Payloads) > 0 {
			payload := op.Payloads[0]

			if err := UnmarshalMap(payload.Properties, &info); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("GetServerInfo() completed without retrieving any data")
		}

		return nil

	})
}

// Retrieve all available sinks from PulseAudio
//
func (self *Client) GetSinks() ([]Sink, error) {
	operation := NewOperation(self)
	defer operation.Destroy()

	sinks := make([]Sink, 0)

	operation.paOper = C.pa_context_get_sink_info_list(self.context, (C.pa_sink_info_cb_t)(unsafe.Pointer(C.pulse_get_sink_info_list_callback)), operation.ToUserdata())

	//  wait for the operation to finish and handle success and error cases
	return sinks, operation.WaitSuccess(func(op *Operation) error {
		//  create a Sink{} for each returned payload
		for _, payload := range op.Payloads {
			sink := Sink{
				Client: self,
			}

			if err := sink.Initialize(payload.Properties); err == nil {
				sinks = append(sinks, sink)
			} else {
				return err
			}
		}

		return nil

	})
}

// Retrieve all available sources from PulseAudio
//
func (self *Client) GetSources() ([]Source, error) {
	operation := NewOperation(self)
	defer operation.Destroy()

	sources := make([]Source, 0)

	operation.paOper = C.pa_context_get_source_info_list(self.context, (C.pa_source_info_cb_t)(unsafe.Pointer(C.pulse_get_source_info_list_callback)), operation.ToUserdata())

	//  wait for the operation to finish and handle success and error cases
	return sources, operation.WaitSuccess(func(op *Operation) error {
		//  create a Source{} for each returned payload
		for _, payload := range op.Payloads {
			source := Source{
				Client: self,
			}

			if err := source.Initialize(payload.Properties); err == nil {
				sources = append(sources, source)
			} else {
				return err
			}
		}

		return nil

	})
}

// Retrieve all available modules from PulseAudio
//
func (self *Client) GetModules() ([]Module, error) {
	operation := NewOperation(self)
	defer operation.Destroy()

	modules := make([]Module, 0)

	operation.paOper = C.pa_context_get_module_info_list(self.context, (C.pa_module_info_cb_t)(unsafe.Pointer(C.pulse_get_module_info_list_callback)), operation.ToUserdata())

	//  wait for the operation to finish and handle success and error cases
	return modules, operation.WaitSuccess(func(op *Operation) error {
		//  create a Module{} for each returned payload
		for _, payload := range op.Payloads {
			module := Module{
				Client: self,
			}

			if err := module.Initialize(payload.Properties); err == nil {
				modules = append(modules, module)
			} else {
				return err
			}
		}

		return nil
	})
}

// Retrieve the last error message from the current context
//
func (self *Client) GetLastError() error {
	if self.context != nil {
		msg := C.GoString(C.pa_strerror(C.pa_context_errno(self.context)))

		if msg != `` {
			return errors.New(msg)
		}
	}

	return nil
}

// Acquire an exclusive lock on the mainloop
//
func (self *Client) Lock() {
	if self.mainloop != nil {
		C.pa_threaded_mainloop_lock(self.mainloop)
	}
}

// Release an exclusive lock on the mainloop
//
func (self *Client) Unlock() {
	if self.mainloop != nil {
		C.pa_threaded_mainloop_unlock(self.mainloop)
	}
}

// Wraps a given function call with a lock
//
func (self *Client) LockFunc(wrapLock ClientLockFunc) error {
	self.Lock()
	err := wrapLock()
	self.Unlock()
	return err
}

// Start the mainloop
//
func (self *Client) Start() error {
	if self.mainloop != nil {
		if status := C.pa_threaded_mainloop_start(self.mainloop); status < 0 {
			return fmt.Errorf("PulseAudio mainloop start failed with code %d", status)
		}
	} else {
		return fmt.Errorf("Cannot operate on undefined PulseAudio mainloop")
	}

	return nil
}

// Wait for a signalling event on the mainloop
//
func (self *Client) Wait() error {
	if self.mainloop != nil {
		C.pa_threaded_mainloop_wait(self.mainloop)
	} else {
		return fmt.Errorf("Cannot operate on undefined PulseAudio mainloop")
	}

	return nil
}

// Send a signalling event to all waiting threads
//
func (self *Client) SignalAll(waitForAccept bool) error {
	if self.mainloop != nil {
		if waitForAccept {
			C.pa_threaded_mainloop_signal(self.mainloop, C.int(1))
		} else {
			C.pa_threaded_mainloop_signal(self.mainloop, C.int(0))
		}
	} else {
		return fmt.Errorf("Cannot operate on undefined PulseAudio mainloop")
	}

	return nil
}

// Stop the mainloop
//
func (self *Client) Stop() error {
	if self.mainloop != nil {
		self.Unlock()
		C.pa_threaded_mainloop_stop(self.mainloop)
	} else {
		return fmt.Errorf("Cannot operate on undefined PulseAudio mainloop")
	}

	return nil
}

// Unregister this client instance from the global CGO tracking pool
//
func (self *Client) Destroy() {
	cgounregister(self.ID)
}

// Wrap this client's ID in a format suitable for passing into C functions as a void-pointer
//
func (self *Client) ToUserdata() unsafe.Pointer {
	return unsafe.Pointer(C.CString(self.ID))
}
