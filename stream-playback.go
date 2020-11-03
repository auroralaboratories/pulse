package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"fmt"
	"io"
	"unsafe"
	// "log"
)

type PlaybackStream struct {
	*Stream
}

func NewPlaybackStream(conn *Conn, name string, sampling *SampleSpec, flags ...StreamFlags) (*PlaybackStream, error) {
	stream := NewStream(conn, name, flags...)

	if err := stream.initialize(); err == nil {
		rv := &PlaybackStream{
			Stream: stream,
		}

		if sampling != nil {
			rv.Sampling = *sampling
		}

		return rv, rv.initialize()
	} else {
		return nil, err
	}
}

func NewPlaybackStreamFromSource(conn *Conn, name string, sampling *SampleSpec, source io.Reader, flags ...StreamFlags) (*PlaybackStream, error) {
	if rv, err := NewPlaybackStream(conn, name, sampling, flags...); err == nil {
		rv.Source = source
		return rv, nil
	} else {
		return nil, err
	}
}

func (self *PlaybackStream) initialize() error {
	if err := self.Stream.initialize(); err != nil {
		return err
	}

	C.pa_stream_set_state_callback(self.Stream.toNative(), (C.pa_stream_notify_cb_t)(C.pulse_stream_state_callback), self.Stream.Userdata())
	C.pa_stream_set_write_callback(self.Stream.toNative(), (C.pa_stream_request_cb_t)(C.pulse_stream_write_callback), self.Stream.Userdata())

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

func Play(conn *Conn, streamName string, sampling *SampleSpec, data io.Reader, flags ...StreamFlags) error {
	if stream, err := NewPlaybackStream(
		conn,
		streamName,
		sampling,
		flags...,
	); err == nil {
		if _, err := io.Copy(stream, data); err == nil {
			if err := stream.Uncork(); err != nil {
				return fmt.Errorf("Failed to uncork stream: %v", err)
			}

			if err := stream.Drain(); err != nil {
				return fmt.Errorf("Failed to drain stream: %v", err)
			}
		} else {
			return err
		}
	} else {
		return fmt.Errorf("Failed to initialize stream: %v", err)
	}

	return nil
}
