package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"io"
	"unsafe"
	// "log"
)

type PlaybackStream struct {
	*Stream
}

func NewPlaybackStream(client *Client, name string) *PlaybackStream {
	return &PlaybackStream{
		Stream: NewStream(client, name),
	}
}

func NewPlaybackStreamFromSource(client *Client, name string, source io.Reader) *PlaybackStream {
	rv := NewPlaybackStream(client, name)
	rv.Source = source
	return rv
}

func (self *PlaybackStream) Initialize() error {
	if err := self.Stream.Initialize(); err != nil {
		return err
	}

	C.pa_stream_set_state_callback(self.Stream.toNative(), (C.pa_stream_notify_cb_t)(C.pulse_stream_state_callback), self.Stream.ToUserdata())
	C.pa_stream_set_write_callback(self.Stream.toNative(), (C.pa_stream_request_cb_t)(C.pulse_stream_write_callback), self.Stream.ToUserdata())

	go func() {
		attr := C.pulse_stream_get_playback_attr(C.int32_t(-1), C.int32_t(-1), C.int32_t(-1), C.int32_t(-1))

		C.pa_stream_connect_playback(self.Stream.toNative(), nil, (*C.pa_buffer_attr)(unsafe.Pointer(&attr)), (C.pa_stream_flags_t)(self.Stream.Flags), nil, nil)
	}()

	//  block until a terminal stream state is reached; successful or otherwise
	select {
	case err := <-self.Stream.state:
		return err
	}
}
