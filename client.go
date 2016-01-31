// +build !cgocheck

/*
Go Bindings for PulseAudio 8.x+
*/
package pulse

// #include "pulse.go.h"
// #cgo pkg-config: libpulse
import "C"

import (
    // "fmt"
    "errors"
    "unsafe"

    log "github.com/Sirupsen/logrus"
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
        log.Warnf("Starting C mainloop")
        C.pulse_mainloop_start(unsafe.Pointer(rv))
        log.Warnf("Completed C mainloop")
    }()

    select {
    case err := <-rv.start:
        if err == nil {
            log.Warnf("Client created and mainloop started")
            return rv, nil
        }else{
            log.Errorf("Client create failed: %v", err)
            return nil, err
        }
    }


    return rv, nil
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


//export go_startPollingOperations
func go_startPollingOperations(clientPtr unsafe.Pointer) {
    // if clientPtr != nil {
    //     client := (*Client)(clientPtr)

    //     go func(){
    //         log.Warnf("Start polling...")

    //         for {
    //             select {
    //             case opCall := <-client.OperationsC:
    //                 log.Warnf("Got op %+v", opCall)
    //                 // opCall.Perform()
    //             }
    //         }
    //     }()
    // }
}
