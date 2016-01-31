// +build !cgocheck
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "unsafe"
    // log "github.com/Sirupsen/logrus"
)

type Client struct {
    Name        string

    start       chan error
}


func NewClient(name string) (*Client, error) {
    rv := &Client{
        Name:        name,
        start:       make(chan error),
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


func (self *Client) GetServerInfo() (ServerInfo, error) {
    operation := NewOperation(self)
    info := ServerInfo{}

    C.pa_context_get_server_info(C.pulse_get_context(), (C.pa_server_info_cb_t)(unsafe.Pointer(C.pulse_get_server_info_callback)), unsafe.Pointer(operation))

//  wait for the operation to finish and handle success and error cases
    return info, operation.Wait(func(op *Operation) error {
        if len(op.Payloads) > 0 {
            payload := op.Payloads[0]

            if err := UnmarshalMap(payload.Properties, &info); err != nil {
                return err
            }
        }else{
            return fmt.Errorf("GetServerInfo() completed without retrieving any data")
        }

        return nil

    }, func(op *Operation, err error) error {
        return err
    })
}


func (self *Client) GetSinks() ([]Sink, error) {
    operation := NewOperation(self)
    sinks := make([]Sink, 0)

    C.pa_context_get_sink_info_list(C.pulse_get_context(), (C.pa_sink_info_cb_t)(unsafe.Pointer(C.pulse_get_sink_info_list_callback)), unsafe.Pointer(operation))

//  wait for the operation to finish and handle success and error cases
    return sinks, operation.Wait(func(op *Operation) error {
    //  create a Sink{} for each returned payload
        for _, payload := range op.Payloads {
            sink := Sink{
                Client: self,
            }

            if err := UnmarshalMap(payload.Properties, &sink); err == nil {
                sink.State = SinkStateInvalid

                if v, ok := payload.Properties[`_state`]; ok {
                    switch v.(type) {
                    case int64:
                        switch int(v.(int64)) {
                        case int(C.PA_SINK_RUNNING):
                            sink.State = SinkStateRunning
                        case int(C.PA_SINK_IDLE):
                            sink.State = SinkStateIdle
                        case int(C.PA_SINK_SUSPENDED):
                            sink.State = SinkStateSuspended
                        }
                    }
                }

                sinks = append(sinks, sink)
            }else{
                return err
            }
        }

        return nil

    }, func(op *Operation, err error) error {
        return err
    })
}