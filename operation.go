package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"sync"
	"time"
	"unsafe"

	"github.com/ghetzel/go-stockutil/stringutil"
)

const (
	DEFAULT_OPERATION_TIMEOUT_MSEC = 5000
)

type OperationSuccessFunc func(*Operation) error
type OperationErrorFunc func(*Operation, error) error

var outstandingOperations sync.Map

// A Payload represents a piece of structured and/or unstructured
// data returned in an Operation. Data that is being retrieved for
// introspection purposes, such as the current state of a sink or
// details about the PulseAudio daemon, will be added to the
// Properties map as key-value pairs.
//
// Other data, such as stream inputs or outputs, will be populated
// in the Data byte slice.
//
type Payload struct {
	Operation  *Operation
	Properties map[string]interface{}
	Data       []byte
}

func NewPayload(operation *Operation) *Payload {
	return &Payload{
		Operation:  operation,
		Properties: make(map[string]interface{}),
		Data:       make([]byte, 0),
	}
}

// An Operation represents a request to a PulseAudio daemon to perform a specific
// task or retrieve data. Operations will either complete successfully (nil will
// be retruned on the Done channel), encounter an error or timeout (non-nil error
// on the Done channel).  Any returned data will be in the form of one or more
// Payload instances in the Payloads slice.
//
type Operation struct {
	ID        string
	Index     int
	Timeout   time.Duration
	Payloads  []*Payload
	lastError error
	paOper    *C.pa_operation
	conn      *Conn
}

func NewOperation(c *Conn) *Operation {
	rv := &Operation{
		ID:       stringutil.UUID().String(),
		conn:     c,
		Index:    -1,
		Timeout:  c.OperationTimeout,
		Payloads: make([]*Payload, 0),
	}

	cgoregister(rv.ID, rv)

	// lock the client for the duration of this operation
	c.Lock()

	return rv
}

// Returns a pointer to the ID used for registering this object
func (self *Operation) Userdata() unsafe.Pointer {
	return unsafe.Pointer(C.CString(self.ID))
}

// Set an error message on this operation
//
func (self *Operation) SetError(err error) {
	self.lastError = err
}

// Retrieve the last error message set on this operation
//
func (self *Operation) GetLastError() error {
	return self.lastError
}

// Performs the operation in a threadsafe manner and returns the last error to occur.
func (self *Operation) Run() error {
	self.conn.Wait()
	err := self.GetLastError()
	self.conn.Unlock()

	return err
}

// Signal the client mainloop that the operation is complete
//
func (self *Operation) Done() {
	self.conn.SignalAll(false)
}

// Create a new payload object and add it to the Payloads stack
//
func (self *Operation) AddPayload() *Payload {
	payload := NewPayload(self)
	self.Payloads = append(self.Payloads, payload)
	return payload
}

// Block the current goroutine until the operation completes, calling the given functions
// on operation success or failure, respectively
//
func (self *Operation) WaitFunc(successFunc OperationSuccessFunc, errorFunc OperationErrorFunc) error {
	// wait for a signalling event from the operations callbacks
	if err := self.Run(); err == nil {
		return successFunc(self)
	} else {
		return errorFunc(self, err)
	}
}

// Block the current goroutine until the operation completes, calling the given
// function if successful.  Errors will pass through and be returned.
//
func (self *Operation) WaitSuccess(successFunc OperationSuccessFunc) error {
	return self.WaitFunc(successFunc, func(op *Operation, err error) error {
		return err
	})
}

// Block the current goroutine until the operation completes, calling the given
// function on failure.  Successful operations will return nil.
//
func (self *Operation) WaitError(errorFunc OperationErrorFunc) error {
	return self.WaitFunc(func(op *Operation) error {
		return nil
	}, errorFunc)
}

// Block the current goroutine until the operation completes.
//
func (self *Operation) Wait() error {
	return self.WaitFunc(func(op *Operation) error {
		return nil
	}, func(op *Operation, err error) error {
		return err
	})
}

func (self *Operation) Destroy() {
	cgounregister(self.ID)
}
