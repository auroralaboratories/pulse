package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "unsafe"
    "github.com/satori/go.uuid"
    "log"
)

// A Stream represents a client-side handle for working with audio data going to or coming from PulseAudio
//
type Stream struct {
    ID                 string
    Client             *Client
    Name               string
    Sampling           SampleSpec

    state              chan error
    paStream           *C.pa_stream
}

func NewStream(client *Client, name string) *Stream {
    rv := &Stream{
        ID:       uuid.NewV4().String(),
        Client:   client,
        Name:     name,
        Sampling: DefaultSampleSpec(),
        state:    make(chan error),
    }

    cgoregister(rv.ID, rv)
    return rv
}

func (self *Stream) Initialize() error {
    spec := (*C.pa_sample_spec)(self.Sampling.ToNative())

    self.paStream = C.pa_stream_new(self.Client.context, C.CString(self.Name), spec, nil)

    return nil
}

func (self *Stream) Uncork() error{
    operation := NewOperation(self.Client)
    operation.Timeout = MaxDuration()
    defer operation.Destroy()

    C.pa_stream_cork(self.ToNative(), C.int(0), (C.pa_stream_success_cb_t)(C.pulse_stream_success_callback), operation.ToUserdata())

    return operation.Wait()
}

func (self *Stream) Cork() error {
    operation := NewOperation(self.Client)
    operation.Timeout = MaxDuration()
    defer operation.Destroy()

    C.pa_stream_cork(self.ToNative(), C.int(1), (C.pa_stream_success_cb_t)(C.pulse_stream_success_callback), operation.ToUserdata())

    return operation.Wait()
}

func (self *Stream) Drain() error {
    operation := NewOperation(self.Client)
    operation.Timeout = MaxDuration()
    defer operation.Destroy()

    C.pa_stream_drain(self.ToNative(), (C.pa_stream_success_cb_t)(C.pulse_stream_success_callback), operation.ToUserdata())

    log.Printf("Waiting for stream %s to drain...", self.Name)

    return operation.Wait()
}


func (self *Stream) ToNative() *C.pa_stream {
    return self.paStream
}

func (self *Stream) Destroy() {
    cgounregister(self.ID)
}

func (self *Stream) ToUserdata() unsafe.Pointer {
    return unsafe.Pointer(C.CString(self.ID))
}
