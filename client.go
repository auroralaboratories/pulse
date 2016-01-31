// +build !cgocheck
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
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
