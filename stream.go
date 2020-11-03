package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "conn.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"bytes"
	"io"
	"log"
	"reflect"
	"unsafe"

	"github.com/ghetzel/go-stockutil/stringutil"
)

const (
	DEFAULT_ASYNC_BUFFER_SIZE = 32768
)

type StreamFlags int

const (
	NoFlags StreamFlags = C.PA_STREAM_NOFLAGS

	// Create the stream corked, requiring an explicit pa_stream_cork() call to uncork it.
	StartCorked = C.PA_STREAM_START_CORKED

	// Interpolate the latency for this stream.
	InterpolateTiming = C.PA_STREAM_INTERPOLATE_TIMING

	// Don't force the time to increase monotonically.
	NotMonotonic = C.PA_STREAM_NOT_MONOTONIC

	// If set timing update requests are issued periodically automatically.
	AutoTimingUpdate = C.PA_STREAM_AUTO_TIMING_UPDATE

	// Don't remap channels by their name, instead map them simply by their index.
	NoRemapChannels = C.PA_STREAM_NO_REMAP_CHANNELS

	// When remapping channels by name, don't upmix or downmix them to related channels.
	NoRemixChannels = C.PA_STREAM_NO_REMIX_CHANNELS

	// Use the sample format of the sink/device this stream is being connected to
	FixFormat = C.PA_STREAM_FIX_FORMAT

	// Use the sample rate of the sink, and possibly ignore the rate the sample spec contains.
	FixRate = C.PA_STREAM_FIX_RATE

	// Use the number of channels and the channel map of the sink
	FixChannels = C.PA_STREAM_FIX_CHANNELS

	// Don't allow moving of this stream to another sink/device.
	DontMove = C.PA_STREAM_DONT_MOVE

	// Allow dynamic changing of the sampling rate during playback with pa_stream_update_sample_rate().
	VariableRate = C.PA_STREAM_VARIABLE_RATE

	// Find peaks instead of resampling.
	PeakDetect = C.PA_STREAM_PEAK_DETECT

	// Create in muted state.
	StartMuted = C.PA_STREAM_START_MUTED

	// Try to adjust the latency of the sink/source based on the requested buffer metrics and adjust buffer metrics accordingly.
	AdjustLatency = C.PA_STREAM_ADJUST_LATENCY

	// Enable compatibility mode for legacy clients that rely on a "classic" hardware device fragment-style playback model.
	EarlyRequests = C.PA_STREAM_EARLY_REQUESTS

	// If set this stream won't be taken into account when it is checked whether the device this stream is connected to should auto-suspend.
	DontInhibitAutoSuspend = C.PA_STREAM_DONT_INHIBIT_AUTO_SUSPEND

	// Create in unmuted state.
	StartUnmuted = C.PA_STREAM_START_UNMUTED

	// If the sink/source this stream is connected to is suspended during the creation of this stream, cause it to fail.
	FailOnSuspend = C.PA_STREAM_FAIL_ON_SUSPEND

	// If a volume is passed when this stream is created, consider it relative to the sink's current volume, never as absolute device volume.
	RelativeVolume = C.PA_STREAM_RELATIVE_VOLUME

	// Used to tag content that will be rendered by passthrough sinks.
	Passthrough = C.PA_STREAM_PASSTHROUGH
)

// A Stream represents a client-side handle for working with audio data going to or coming from PulseAudio
//
type Stream struct {
	BufferSize  int
	Destination io.Writer
	Flags       StreamFlags
	ID          string
	Name        string
	Sampling    SampleSpec
	Source      io.Reader
	state       chan error
	paStream    *C.pa_stream
	buffer      *bytes.Buffer
	conn        *Conn
}

func NewStream(conn *Conn, name string, flags ...StreamFlags) *Stream {
	rv := &Stream{
		BufferSize: DEFAULT_ASYNC_BUFFER_SIZE,
		conn:       conn,
		ID:         stringutil.UUID().String(),
		Name:       name,
		Sampling:   DefaultSampleSpec(),
		Flags:      NoFlags,

		state: make(chan error),
	}

	if len(flags) > 0 {
		rv.AddFlags(flags...)
	}

	cgoregister(rv.ID, rv)
	return rv
}

func (self *Stream) initialize() error {
	spec := (*C.pa_sample_spec)(self.Sampling.toNative())

	self.buffer = bytes.NewBuffer(make([]byte, 0, self.BufferSize))

	if self.Source == nil {
		self.Source = self.buffer
	}

	// create the client-side stream object
	self.paStream = C.pa_stream_new(
		self.conn.context,
		C.CString(self.Name),
		spec,
		nil,
	)

	return nil
}

func (self *Stream) AddFlags(flags ...StreamFlags) {
	for _, flag := range flags {
		self.Flags |= flag
	}
}

// Return whether the current stream is corked (stopped) or not
//
func (self *Stream) IsCorked() bool {
	return (int(C.pa_stream_is_corked(self.toNative())) == 0)
}

// Uncork (start) the stream
//
func (self *Stream) Uncork() error {
	if self.IsCorked() {
		operation := NewOperation(self.conn)
		operation.Timeout = MaxDuration()
		defer operation.Destroy()

		operation.paOper = C.pa_stream_cork(
			self.toNative(),
			C.int(0),
			(C.pa_stream_success_cb_t)(C.pulse_stream_success_callback),
			operation.Userdata(),
		)

		return operation.WaitSuccess(func(op *Operation) error {
			log.Printf("Waiting for stream %s uncorked", self.Name)
			return nil
		})
	} else {
		return nil
	}
}

// Cork (stop) the stream
//
func (self *Stream) Cork() error {
	if !self.IsCorked() {
		operation := NewOperation(self.conn)
		operation.Timeout = MaxDuration()
		defer operation.Destroy()

		operation.paOper = C.pa_stream_cork(
			self.toNative(),
			C.int(1),
			(C.pa_stream_success_cb_t)(C.pulse_stream_success_callback),
			operation.Userdata(),
		)

		return operation.WaitSuccess(func(op *Operation) error {
			log.Printf("Waiting for stream %s corked", self.Name)
			return nil
		})
	} else {
		return nil
	}
}

// Block until the stream's buffer has fully played
//
func (self *Stream) Drain() error {
	operation := NewOperation(self.conn)
	operation.Timeout = MaxDuration()
	defer operation.Destroy()

	operation.paOper = C.pa_stream_drain(
		self.toNative(),
		(C.pa_stream_success_cb_t)(C.pulse_stream_success_callback),
		operation.Userdata(),
	)

	return operation.WaitSuccess(func(op *Operation) error {
		log.Printf("Waiting for stream %s drained", self.Name)
		return nil
	})
}

// func (self *Stream) Read(data []byte) (int, error) {
//    return self.buffer.Read(data)
// }

func (self *Stream) Write(data []byte) (int, error) {
	return self.buffer.Write(data)
}

// Return the stream's native C pointer
//
func (self *Stream) toNative() *C.pa_stream {
	return self.paStream
}

func (self *Stream) Destroy() {
	if p := self.toNative(); p != nil {
		C.pa_stream_disconnect(p)
	}

	cgounregister(self.ID)
}

func (self *Stream) Userdata() unsafe.Pointer {
	return unsafe.Pointer(C.CString(self.ID))
}

func (self *Stream) readFromSource(length int) {
	if self.Source != nil {
		// allocate C buffer
		cData := C.malloc(C.size_t(length + 1))
		toFill := C.size_t(length)

		// call begin write to determine how much data the server wants
		if status := int(C.pa_stream_begin_write(self.toNative(), &cData, &toFill)); status < 0 {
			// return -1, self.conn.GetLastError()
			log.Printf("Write Prep failed: %d", status)
			return
		}

		// allocate local byteslice to the size of the server's choosing
		data := make([]byte, int(toFill))

		// read data from the source
		if n, err := self.Source.Read(data); err == nil {
			gSlice := &reflect.SliceHeader{
				Data: uintptr(cData),
				Len:  n,
				Cap:  n,
			}

			c := *(*[]byte)(unsafe.Pointer(gSlice))

			copy(c, data)

			// perform the PulseAudio write operation
			if status := int(C.pa_stream_write(self.toNative(), unsafe.Pointer(cData), toFill, nil, 0, C.PA_SEEK_RELATIVE)); status < 0 {
				// return -1, io.ErrUnexpectedEOF
				log.Printf("Write failed (%d): %v", status, self.conn.GetLastError())
				return
			} else {
				// log.Printf("pulse.Stream(%s).Write(%d bytes); wrote %d, status=%d\n", self.ID, n, int(toFill), status)
				return
			}
		}
	}
}

// func (self *Stream) writeNFromBuffer(length int) {
//    bytes_remaining := length
//    bytesWritten := 0

//    for bytes_remaining > 0 {
//        log.Printf("Buffer len %d\n", self.buffer.Len())

//        data := make([]byte, length)

//    // read `length' bytes from the stream buffer
//        if n, err := self.buffer.Read(data); err == nil {
//            log.Printf("Write %d/%d bytes from internal buffer %s (size: %d)\n", n, length, self.ID, self.buffer.Len())

//        // only do the complicated stuff if there was any data in there
//            if n > 0 {

//            }
//        }else if err == io.EOF {
//            status <- nil
//        }else{
//            time.Sleep(500 * time.Millisecond)
//        }
//    }
// }
