package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "time"
    // log "github.com/Sirupsen/logrus"
)

const (
    DEFAULT_OPERATION_TIMEOUT_MSEC = 5000
)

type OperationSuccessFunc func(*Operation) error
type OperationErrorFunc   func(*Operation, error) error

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
    Client     *Client
    Done       chan error
    Index      int
    Timeout    time.Duration
    Payloads   []*Payload

    paOper     *C.pa_operation
}

func NewOperation(client *Client) *Operation {
    return &Operation{
        Client:     client,
        Done:       make(chan error),
        Index:      -1,
        Timeout:    client.OperationTimeout,
        Payloads:   make([]*Payload, 0),
    }
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
    select{
    case err := <- self.Done:
        if err == nil {
            return successFunc(self)
        }else{
            return errorFunc(self, err)
        }
    case <-time.After(self.Timeout):
        return errorFunc(self, fmt.Errorf("Timed out waiting for operation to complete (timeout: %s)", self.Timeout))
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