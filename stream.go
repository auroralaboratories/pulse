package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "bytes"
    "io"
    "log"
    // "time"
    "unsafe"
    "reflect"

    "github.com/satori/go.uuid"
)

const (
    DEFAULT_ASYNC_BUFFER_SIZE = 32768
)

// A Stream represents a client-side handle for working with audio data going to or coming from PulseAudio
//
type Stream struct {
    ID                 string
    Client             *Client
    Name               string
    Sampling           SampleSpec
    BufferSize         int

    state              chan error
    paStream           *C.pa_stream
    buffer             *bytes.Buffer

    Source             io.Reader
    Destination        io.Writer
}

func NewStream(client *Client, name string) *Stream {
    rv := &Stream{
        BufferSize: DEFAULT_ASYNC_BUFFER_SIZE,
        Client:     client,
        ID:         uuid.NewV4().String(),
        Name:       name,
        Sampling:   DefaultSampleSpec(),

        state:      make(chan error),
    }

    cgoregister(rv.ID, rv)
    return rv
}

func (self *Stream) Initialize() error {
    spec := (*C.pa_sample_spec)(self.Sampling.toNative())

//  initialize I/O buffer
    self.buffer   = bytes.NewBuffer(make([]byte, 0, self.BufferSize))

//  create the client-side stream object
    self.paStream = C.pa_stream_new(self.Client.context, C.CString(self.Name), spec, nil)

    return nil
}

// Return whether the current stream is corked (stopped) or not
//
func (self *Stream) IsCorked() bool {
    return (int(C.pa_stream_is_corked(self.toNative())) == 0)
}


// Uncork (start) the stream
//
func (self *Stream) Uncork() error{
    operation := NewOperation(self.Client)
    operation.Timeout = MaxDuration()
    defer operation.Destroy()

    operation.paOper = C.pa_stream_cork(self.toNative(), C.int(0), (C.pa_stream_success_cb_t)(C.pulse_stream_success_callback), operation.ToUserdata())

    return operation.WaitSuccess(func(op *Operation) error {
        log.Printf("Waiting for stream %s uncorked", self.Name)
        return nil
    })
}


// Cork (stop) the stream
//
func (self *Stream) Cork() error {
    operation := NewOperation(self.Client)
    operation.Timeout = MaxDuration()
    defer operation.Destroy()

    operation.paOper = C.pa_stream_cork(self.toNative(), C.int(1), (C.pa_stream_success_cb_t)(C.pulse_stream_success_callback), operation.ToUserdata())

    return operation.WaitSuccess(func(op *Operation) error {
        log.Printf("Waiting for stream %s corked", self.Name)
        return nil
    })
}


// Block until the stream's buffer has fully played
//
func (self *Stream) Drain() error {
    operation := NewOperation(self.Client)
    operation.Timeout = MaxDuration()
    defer operation.Destroy()

    operation.paOper = C.pa_stream_drain(self.toNative(), (C.pa_stream_success_cb_t)(C.pulse_stream_success_callback), operation.ToUserdata())

    return operation.WaitSuccess(func(op *Operation) error {
        log.Printf("Waiting for stream %s drained", self.Name)
        return nil
    })
}

// func (self *Stream) Read(data []byte) (int, error) {
//     return self.buffer.Read(data)
// }

// func (self *Stream) Write(data []byte) (int, error) {

// }

// Return the stream's native C pointer
//
func (self *Stream) toNative() *C.pa_stream {
    return self.paStream
}

func (self *Stream) Destroy() {
    cgounregister(self.ID)
}

func (self *Stream) ToUserdata() unsafe.Pointer {
    return unsafe.Pointer(C.CString(self.ID))
}

func (self *Stream) readFromSource(length int) {

    if self.Source != nil {
        cData  := C.malloc(C.size_t(length+1))
        toFill := C.size_t(length)

        if status := int(C.pa_stream_begin_write(self.toNative(), &cData, &toFill)); status < 0 {
            // return -1, self.Client.GetLastError()
            log.Printf("Write Prep failed: %d", status)
            return
        }

        data := make([]byte, int(toFill))

        if n, err := self.Source.Read(data); err == nil {
            log.Printf("[%s] read %d bytes from source %+v; got %d", self.ID, int(toFill), self.Source, n)

            gSlice := &reflect.SliceHeader{
                Data: uintptr(cData),
                Len:  n,
                Cap:  n,
            }

            c := *(*[]byte)(unsafe.Pointer(gSlice))

            cn := copy(c, data)
            log.Printf("Copied %d from data -> c\n", cn)

        //  perform the PulseAudio write operation
            if status := int(C.pa_stream_write(self.toNative(), unsafe.Pointer(cData), toFill, nil, 0, C.PA_SEEK_RELATIVE)); status < 0 {
                // return -1, io.ErrUnexpectedEOF
                log.Printf("Write failed (%d): %v", status, self.Client.GetLastError())
                return
            }else{
                log.Printf("pulse.Stream(%s).Write(%d bytes); wrote %d, status=%d\n", self.ID, n, int(toFill), status)
                return
            }
        }
    }
}

// func (self *Stream) writeNFromBuffer(length int) {
//     bytes_remaining := length
//     bytesWritten := 0

//     for bytes_remaining > 0 {
//         log.Printf("Buffer len %d\n", self.buffer.Len())

//         data := make([]byte, length)

//     //  read `length' bytes from the stream buffer
//         if n, err := self.buffer.Read(data); err == nil {
//             log.Printf("Write %d/%d bytes from internal buffer %s (size: %d)\n", n, length, self.ID, self.buffer.Len())

//         //  only do the complicated stuff if there was any data in there
//             if n > 0 {

//             }
//         }else if err == io.EOF {
//             status <- nil
//         }else{
//             time.Sleep(500 * time.Millisecond)
//         }
//     }
// }
