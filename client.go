// +build !cgocheck

/*
Go Bindings for PulseAudio 8.x+
*/
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "errors"
    "unsafe"

    "github.com/shutterstock/go-stockutil/stringutil"
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
    rv := ServerInfo{}

    C.pa_context_get_server_info(C.pulse_get_context(), (C.pa_server_info_cb_t)(unsafe.Pointer(C.pulse_get_server_info_callback)), unsafe.Pointer(operation))

    select{
    case err := <- operation.Done:
        if err == nil {
            if err := UnmarshalMap(operation.Properties, &rv); err == nil {
                return rv, nil
            }else{
                return rv, err
            }
        }else{
            return rv, err
        }
    }
}

//export go_clientStartupDone
func go_clientStartupDone(clientPtr unsafe.Pointer, message *C.char) {
    if clientPtr != nil {
        client := (*Client)(clientPtr)

        if str := C.GoString(message); str == `` {
            client.start <- nil
        }else{
            client.start <- errors.New(str)
        }
    }
}


//export go_operationSetProperty
func go_operationSetProperty(operationPtr unsafe.Pointer, k *C.char, v *C.char, convertTo *C.char) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)

        if key := C.GoString(k); key != `` {
            if value := C.GoString(v); value != `` {
                if convertTo != nil {
                    var ctype stringutil.ConvertType

                    switch C.GoString(convertTo) {
                    case `bool`:
                        ctype = stringutil.Boolean
                    case `float`:
                        ctype = stringutil.Float
                    case `int`:
                        ctype = stringutil.Integer
                    case `time`:
                        ctype = stringutil.Time
                    default:
                        ctype = stringutil.String
                    }

                    if convertedValue, err := stringutil.ConvertTo(ctype, value); err == nil {
                        operation.Properties[key] = convertedValue
                    }else{
                        operation.Properties[key] = value
                    }
                }else{
                    operation.Properties[key] = value
                }
            }
        }
    }
}


//export go_operationComplete
func go_operationComplete(operationPtr unsafe.Pointer) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)
        operation.Done <- nil
    }
}


//export go_operationFailed
func go_operationFailed(operationPtr unsafe.Pointer, message *C.char) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)

        if msg := C.GoString(message); msg == `` {
            operation.Done <- fmt.Errorf("Unknown error")
        }else{
            operation.Done <- errors.New(msg)
        }
    }
}