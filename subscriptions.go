// Golang bindings for PulseAudio 8.x+
package pulse

// #cgo CFLAGS: -Wno-error=implicit-function-declaration
// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"unsafe"
)

type EventType int

const (
	NullEvent         EventType = C.PA_SUBSCRIPTION_MASK_NULL          // No events.
	SinkEvent                   = C.PA_SUBSCRIPTION_MASK_SINK          // Sink events.
	SourceEvent                 = C.PA_SUBSCRIPTION_MASK_SOURCE        // Source events.
	SinkInputEvent              = C.PA_SUBSCRIPTION_MASK_SINK_INPUT    // Sink input events.
	SourceOutputEvent           = C.PA_SUBSCRIPTION_MASK_SOURCE_OUTPUT // Source output events.
	ModuleEvent                 = C.PA_SUBSCRIPTION_MASK_MODULE        // Module events.
	ClientEvent                 = C.PA_SUBSCRIPTION_MASK_CLIENT        // Client events.
	SampleCacheEvent            = C.PA_SUBSCRIPTION_MASK_SAMPLE_CACHE  // Sample cache events.
	ServerEvent                 = C.PA_SUBSCRIPTION_MASK_SERVER        // Other global server changes.
	CardEvent                   = C.PA_SUBSCRIPTION_MASK_CARD          // Card events.
	AllEvent                    = C.PA_SUBSCRIPTION_MASK_ALL           // Catch all events.
)

func ExtractEvents(combined int) []EventType {
	types := make([]EventType, 0)

	for _, eventType := range []EventType{
		SinkEvent,
		SourceEvent,
		SinkInputEvent,
		SourceOutputEvent,
		ModuleEvent,
		ClientEvent,
		SampleCacheEvent,
		ServerEvent,
		CardEvent,
	} {
		if combined&int(eventType) == int(eventType) {
			types = append(types, eventType)
		}
	}

	return types
}

func (self EventType) String() string {
	switch self {
	case SinkEvent:
		return `sink`
	case SourceEvent:
		return `source`
	case SinkInputEvent:
		return `sink-input`
	case SourceOutputEvent:
		return `source-output`
	case ModuleEvent:
		return `module`
	case ClientEvent:
		return `client`
	case SampleCacheEvent:
		return `sample-cache`
	case ServerEvent:
		return `server`
	case CardEvent:
		return `card`
	default:
		return `unknown`
	}
}

var eventTypes chan EventType

// Subscribe to event notifications and emit the type of event as it occurs.
func (self *Client) Subscribe(types ...EventType) <-chan EventType {
	eventTypes = make(chan EventType)

	// subscribe to event types
	operation := NewOperation(self)
	defer operation.Destroy()

	typeMask := 0

	if len(types) == 0 {
		typeMask = AllEvent
	} else {
		for _, tm := range types {
			typeMask |= int(tm)
		}
	}

	operation.paOper = C.pa_context_subscribe(
		self.context,
		(C.pa_subscription_mask_t)(C.int(typeMask)),
		(C.pa_context_success_cb_t)(C.pulse_generic_success_callback),
		operation.Userdata(),
	)

	// wait for the result
	err := operation.Wait()

	if err != nil {
		close(eventTypes)
	}

	// set subscription callback
	C.pa_context_set_subscribe_callback(
		self.context,
		(C.pa_context_subscribe_cb_t)(
			unsafe.Pointer(C.pulse_subscription_event_callback),
		),
		self.Userdata(),
	)

	return eventTypes
}

//export go_clientEventCallback
func go_clientEventCallback(types C.pa_subscription_event_type_t, index C.uint32_t, clientId *C.char) {
	if _, ok := cgoget(C.GoString(clientId)).(*Client); ok {
		for _, eventType := range ExtractEvents(int(types)) {
			eventTypes <- eventType
		}
	}
}
