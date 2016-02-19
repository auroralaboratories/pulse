package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    // "fmt"
    "io"
    "unsafe"
    "log"
    "reflect"
)

type PlaybackStream struct {
    *Stream
    io.Writer
}

func NewPlaybackStream(client *Client, name string) *PlaybackStream {
    return &PlaybackStream{
        Stream: NewStream(client, name),
    }
}

func (self *PlaybackStream) Initialize() error {
    if err := self.Stream.Initialize(); err != nil {
        return err
    }

    paStream := (*C.pa_stream)(self.Stream.ToNative())

    C.pa_stream_set_state_callback(paStream, (C.pa_stream_notify_cb_t)(C.pulse_stream_state_callback), unsafe.Pointer(self.Stream))

    go func(){
        C.pa_stream_connect_playback(paStream, nil, nil, (C.pa_stream_flags_t)(0), nil, nil)
    }()

    select {
    case err := <-self.Stream.state:
        return err
    }
}

func (self *PlaybackStream) Write(data []byte) (int, error) {
    log.Printf("Write: %d\n", len(data))

    paStream := (*C.pa_stream)(self.Stream.ToNative())

    dataTyp := reflect.TypeOf(data)
    dataVal := reflect.ValueOf(data)
    dataLen := dataVal.Len()
    dataSz  := C.size_t(dataTyp.Elem().Size() * uintptr(dataLen))

    if dataLen > 0 {
        written := C.pulse_stream_write(paStream, unsafe.Pointer(dataVal.Index(0).UnsafeAddr()), dataSz, nil)

        if int(written) < 0 {
            return -1, io.ErrUnexpectedEOF
        }

        return int(written), nil
    }

    return 0, io.ErrUnexpectedEOF
}