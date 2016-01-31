// +build !cgocheck
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

type Operation struct {
    Client     *Client
    Done       chan error
    Index      int
    Timeout    time.Duration
    Payloads   []*Payload
}

func NewOperation(client *Client) *Operation {
    return &Operation{
        Client:     client,
        Done:       make(chan error),
        Index:      -1,
        Timeout:    (time.Duration(DEFAULT_OPERATION_TIMEOUT_MSEC) * time.Millisecond),
        Payloads:   make([]*Payload, 0),
    }
}

func (self *Operation) AddPayload() *Payload {
    payload := NewPayload(self)
    self.Payloads = append(self.Payloads, payload)
    return payload
}

func (self *Operation) Wait(successFunc OperationSuccessFunc, errorFunc OperationErrorFunc) error {
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