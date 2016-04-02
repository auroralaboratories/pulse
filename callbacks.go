package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
	"errors"
	"fmt"
	// "unsafe"

	"github.com/shutterstock/go-stockutil/stringutil"
	// log "github.com/Sirupsen/logrus"
)

//export go_clientStartupDone
func go_clientStartupDone(clientId *C.char, message *C.char) {
	if obj, ok := cgoget(C.GoString(clientId)); ok {
		switch obj.(type) {
		case *Client:
			client := obj.(*Client)
			client.SignalAll(false)
		default:
			panic(fmt.Sprintf("go_clientStartupDone(): invalid object %s; expected *pulse.Client, got %T", clientId, obj))
		}
	}
}

//export go_operationSetProperty
func go_operationSetProperty(operationId *C.char, k *C.char, v *C.char, convertTo *C.char) {
	if obj, ok := cgoget(C.GoString(operationId)); ok {
		switch obj.(type) {
		case *Operation:
			operation := obj.(*Operation)

			var payload *Payload

			if operation.Index < 0 {
				payload = operation.AddPayload()
				operation.Index = 0
			} else {
				payload = operation.Payloads[operation.Index]
			}

			if key := C.GoString(k); key != `` {
				if value := C.GoString(v); value != `` {
					if convertTo != nil {
						var ctype stringutil.ConvertType

						switch C.GoString(convertTo) {
						case `bool`:
							ctype = stringutil.Boolean
						case `float`:
							ctype = stringutil.Float
						case `int`:
							ctype = stringutil.Integer
						case `time`:
							ctype = stringutil.Time
						default:
							ctype = stringutil.String
						}

						if convertedValue, err := stringutil.ConvertTo(ctype, value); err == nil {
							payload.Properties[key] = convertedValue
						} else {
							payload.Properties[key] = value
						}
					} else {
						payload.Properties[key] = value
					}
				}
			}
		default:
			panic(fmt.Sprintf("go_operationSetProperty(): invalid object %s; expected *pulse.Operation, got %T", operationId, obj))
		}
	}
}

//export go_operationCreatePayload
func go_operationCreatePayload(operationId *C.char) {
	if obj, ok := cgoget(C.GoString(operationId)); ok {
		switch obj.(type) {
		case *Operation:
			operation := obj.(*Operation)
			operation.AddPayload()
			operation.Index = (len(operation.Payloads) - 1)
		default:
			panic(fmt.Sprintf("go_operationCreatePayload(): invalid object %s; expected *pulse.Operation, got %T", operationId, obj))
		}
	}
}

//export go_operationComplete
func go_operationComplete(operationId *C.char) {
	if obj, ok := cgoget(C.GoString(operationId)); ok {
		switch obj.(type) {
		case *Operation:
			operation := obj.(*Operation)

			//  truncate empty payloads
			for i, payload := range operation.Payloads {
				if len(payload.Properties) == 0 && len(payload.Data) == 0 {
					operation.Payloads = append(operation.Payloads[:i], operation.Payloads[i+1:]...)
				}
			}

			//  unref pa_operation
			if operation.paOper != nil {
				C.pa_operation_unref(operation.paOper)
			}

			operation.Done()
		default:
			panic(fmt.Sprintf("go_operationComplete(): invalid object %s; expected *pulse.Operation, got %T", operationId, obj))
		}
	}
}

//export go_operationFailed
func go_operationFailed(operationId *C.char, message *C.char) {
	if obj, ok := cgoget(C.GoString(operationId)); ok {
		switch obj.(type) {
		case *Operation:
			operation := obj.(*Operation)

			//  unref pa_operation
			if operation.paOper != nil {
				C.pa_operation_unref(operation.paOper)
			}

			if msg := C.GoString(message); msg == `` {
				operation.SetError(fmt.Errorf("Unknown error"))
			} else {
				operation.SetError(errors.New(msg))
			}

			operation.Done()
		default:
			panic(fmt.Sprintf("go_operationFailed(): invalid object %s; expected *pulse.Operation, got %T", operationId, obj))
		}
	}
}
