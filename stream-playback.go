package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    // "fmt"
    "io"
    "unsafe"
    // "log"
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

    C.pa_stream_set_state_callback(self.Stream.ToNative(), (C.pa_stream_notify_cb_t)(C.pulse_stream_state_callback), self.Stream.ToUserdata())

    go func(){
        C.pa_stream_connect_playback(self.Stream.ToNative(), nil, nil, (C.pa_stream_flags_t)(0), nil, nil)
    }()

    select {
    case err := <-self.Stream.state:
        return err
    }
}

func (self *PlaybackStream) IsPlaying() bool {
    return (int(C.pa_stream_is_corked(self.Stream.ToNative())) == 0)
}

func (self *PlaybackStream) Write(data []byte) (int, error) {
    cData := C.malloc(C.size_t(len(data)+1))
    dest := (*[1<<30]byte)(cData)
    copy(dest[:], data)
    dest[len(data)] = 0

    dataTyp := reflect.TypeOf(data)
    dataSz  := C.size_t(dataTyp.Elem().Size() * uintptr(len(data)))

    if len(data) > 0 {
        written := C.pulse_stream_write(self.Stream.ToNative(), unsafe.Pointer((*C.uint8_t)(cData)), dataSz, nil)

        if int(written) < 0 {
            return -1, io.ErrUnexpectedEOF
        }

        return int(written), nil
    }

    return 0, io.ErrUnexpectedEOF
}
