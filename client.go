// Golang bindings for PulseAudio 8.x+
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "time"
    "unsafe"
    // log "github.com/Sirupsen/logrus"
)

// A PulseAudio Client represents a connection to a PulseAudio daemon (either locally or
// on a remote host). A Client is the primary entry point for working with PulseAudio
// objects and data.
//
type Client struct {
    Name             string
    OperationTimeout time.Duration

    start            chan error
}


func NewClient(name string) (*Client, error) {
    rv := &Client{
        Name:             name,
        OperationTimeout: (time.Duration(DEFAULT_OPERATION_TIMEOUT_MSEC) * time.Millisecond),

        start:            make(chan error),
    }

    go func(){
        C.pulse_mainloop_start(C.CString(name), unsafe.Pointer(rv))
    }()

    select {
    case err := <-rv.start:
        if err == nil {
            return rv, nil
        }else{
            return nil, err
        }
    }


    return rv, nil
}


// Retrieve information about the connected PulseAudio daemon
//
func (self *Client) GetServerInfo() (ServerInfo, error) {
    operation := NewOperation(self)
    info := ServerInfo{}

    operation.paOper = C.pa_context_get_server_info(C.pulse_get_context(), (C.pa_server_info_cb_t)(unsafe.Pointer(C.pulse_get_server_info_callback)), unsafe.Pointer(operation))

//  wait for the operation to finish and handle success and error cases
    return info, operation.WaitSuccess(func(op *Operation) error {
        if len(op.Payloads) > 0 {
            payload := op.Payloads[0]

            if err := UnmarshalMap(payload.Properties, &info); err != nil {
                return err
            }
        }else{
            return fmt.Errorf("GetServerInfo() completed without retrieving any data")
        }

        return nil

    })
}


// Retrieve all available sinks from PulseAudio
//
func (self *Client) GetSinks() ([]Sink, error) {
    operation := NewOperation(self)
    sinks := make([]Sink, 0)

    operation.paOper = C.pa_context_get_sink_info_list(C.pulse_get_context(), (C.pa_sink_info_cb_t)(unsafe.Pointer(C.pulse_get_sink_info_list_callback)), unsafe.Pointer(operation))

//  wait for the operation to finish and handle success and error cases
    return sinks, operation.WaitSuccess(func(op *Operation) error {
    //  create a Sink{} for each returned payload
        for _, payload := range op.Payloads {
            sink := Sink{
                Client: self,
            }

            if err := sink.Initialize(payload.Properties); err == nil {
                sinks = append(sinks, sink)
            }else{
                return err
            }
        }

        return nil

    })
}