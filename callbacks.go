package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
    "errors"
    "unsafe"

    "github.com/shutterstock/go-stockutil/stringutil"
    // log "github.com/Sirupsen/logrus"
)


//export go_clientStartupDone
func go_clientStartupDone(clientPtr unsafe.Pointer, message *C.char) {
    if clientPtr != nil {
        client := (*Client)(clientPtr)

        if str := C.GoString(message); str == `` {
            client.start <- nil
        }else{
            client.start <- errors.New(str)
        }
    }
}


//export go_operationSetProperty
func go_operationSetProperty(operationPtr unsafe.Pointer, k *C.char, v *C.char, convertTo *C.char) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)
        var payload *Payload

        if operation.Index < 0 {
            payload = operation.AddPayload()
            operation.Index = 0
        }else{
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
                    }else{
                        payload.Properties[key] = value
                    }
                }else{
                    payload.Properties[key] = value
                }
            }
        }
    }
}

//export go_operationCreatePayload
func go_operationCreatePayload(operationPtr unsafe.Pointer) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)
        operation.AddPayload()
        operation.Index = (len(operation.Payloads) - 1)
    }
}


//export go_operationComplete
func go_operationComplete(operationPtr unsafe.Pointer) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)

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

        operation.Done <- nil
    }
}


//export go_operationFailed
func go_operationFailed(operationPtr unsafe.Pointer, message *C.char) {
    if operationPtr != nil {
        operation := (*Operation)(operationPtr)

    //  unref pa_operation
        if operation.paOper != nil {
            C.pa_operation_unref(operation.paOper)
        }

        if msg := C.GoString(message); msg == `` {
            operation.Done <- fmt.Errorf("Unknown error")
        }else{
            operation.Done <- errors.New(msg)
        }
    }
}