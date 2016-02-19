// Golang bindings for PulseAudio 8.x+
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "time"
    "unsafe"
    "github.com/satori/go.uuid"
    // log "github.com/Sirupsen/logrus"
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

    state            chan error
    mainloop         *C.pa_mainloop
    context          *C.pa_context
    api              *C.pa_mainloop_api

}


func NewClient(name string) (*Client, error) {
    rv := &Client{
        ID:               uuid.NewV4().String(),
        Name:             name,
        OperationTimeout: (time.Duration(DEFAULT_OPERATION_TIMEOUT_MSEC) * time.Millisecond),

        state:            make(chan error),
    }

    cgoregister(rv.ID, rv)

    go func(){
        userdata := C.CString(rv.ID)

        rv.mainloop = C.pa_mainloop_new()
        if rv.mainloop == nil {
            go_clientStartupDone(userdata, C.CString("Failed to create PulseAudio mainloop"))
            return
        }

        rv.api     = C.pa_mainloop_get_api(rv.mainloop)
        rv.context = C.pa_context_new(rv.api, C.CString(name))

        C.pa_context_set_state_callback(rv.context, (C.pa_context_notify_cb_t)(C.pulse_context_state_callback), rv.ToUserdata())

    //  being context connect
        if int(C.pa_context_connect(rv.context, nil, (C.pa_context_flags_t)(0), nil)) < 0 {
            msg := fmt.Sprintf("Failed to connect PulseAudio context: %s", C.GoString(C.pa_strerror(C.pa_context_errno(rv.context))))

            go_clientStartupDone(userdata, C.CString(msg))
            return;
        }

    //  start pulseaudio mainloop
        if int(C.pa_mainloop_run(rv.mainloop, nil)) < 0 {
            go_clientStartupDone(userdata, C.CString("Failed to start PulseAudio mainloop"));
            return;
        }
    }()

    select {
    case err := <-rv.state:
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
            }else{
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
            }else{
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
            }else{
                return err
            }
        }

        return nil
    })
}

func (self *Client) Destroy() {
    cgounregister(self.ID)
}

func (self *Client) ToUserdata() unsafe.Pointer {
    return unsafe.Pointer(C.CString(self.ID))
}
