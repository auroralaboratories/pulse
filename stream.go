package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "unsafe"
)


// A Stream represents a client-side handle for working with audio data going to or coming from PulseAudio
//
type Stream struct {
    Client             *Client
    Name               string
    Sampling           SampleSpec

    state              chan error
    paStream           unsafe.Pointer
}

func NewStream(client *Client, name string) *Stream {
    return &Stream{
        Client:   client,
        Name:     name,
        Sampling: DefaultSampleSpec(),
        state:    make(chan error),
    }
}

func (self *Stream) Initialize() error {
    spec := (*C.pa_sample_spec)(self.Sampling.ToNative())

    stream := C.pa_stream_new(C.pulse_get_context(), C.CString(self.Name), spec, nil)
    self.paStream = unsafe.Pointer(stream)

    return nil
}

func (self *Stream) ToNative() unsafe.Pointer {
    return self.paStream
}