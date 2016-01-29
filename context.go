// +build !cgocheck

/*
Go Bindings for PulseAudio 8.x+
*/
package pulse

// #include "pulse.go.h"
// #cgo pkg-config: libpulse
import "C"

import (
    "fmt"
)

type ContextFlags int
const (
    CONTEXT_NOFLAGS     ContextFlags = C.PA_CONTEXT_NOFLAGS
    CONTEXT_NOAUTOSPAWN              = C.PA_CONTEXT_NOAUTOSPAWN
    CONTEXT_NOFAIL                   = C.PA_CONTEXT_NOFAIL
)

type ContextState int
const (
    CONTEXT_UNCONNECTED ContextState = C.PA_CONTEXT_UNCONNECTED
    CONTEXT_CONNECTING               = C.PA_CONTEXT_CONNECTING
    CONTEXT_AUTHORIZING              = C.PA_CONTEXT_AUTHORIZING
    CONTEXT_SETTING_NAME             = C.PA_CONTEXT_SETTING_NAME
    CONTEXT_READY                    = C.PA_CONTEXT_READY
    CONTEXT_FAILED                   = C.PA_CONTEXT_FAILED
    CONTEXT_TERMINATED               = C.PA_CONTEXT_TERMINATED
)


type Context struct {
    Client  *Client
    Name    string

    context *C.pa_context
}


func (self *Context) Initialize(flags ContextFlags) error {
    self.context = C.pa_context_new(self.Client.api, C.CString(self.Name))

    if self.Client.mainloop == nil {
        return fmt.Errorf("Uninitialized PulseAudio mainloop")
    }

    if self.Client.api == nil {
        return fmt.Errorf("Uninitialized PulseAudio mainloop API")
    }

    //  register context state change callback
    //  pa_context_set_state_callback(self.)


    self.Client.Lock()

    if i := int(C.pa_context_connect(self.context, nil, C.pa_context_flags_t(int(flags)), nil)); i < 0 {
        return fmt.Errorf("Failed to connect to PulseAudio server")
    }

    if err := self.WaitUntil(CONTEXT_READY); err != nil {
        defer self.Destroy()
        return err
    }

    self.Client.Unlock()

    return nil
}

func (self *Context) WaitUntil(state ContextState) error {
    if err := self.Client.Lock(); err != nil {
        return err
    }

    for {
        switch int(C.pa_context_get_state(self.context)) {
        case int(state):
            return nil
        case int(CONTEXT_FAILED):
            return fmt.Errorf("Context entered a failed state waiting to enter state %d", state)
        case int(CONTEXT_TERMINATED):
            return fmt.Errorf("Context was terminated waiting to enter state %d", state)
        default:
            self.Client.Wait()
        }
    }

    return self.Client.Unlock()
}

func (self *Context) Destroy() {
    if self.context != nil {
        C.pa_context_unref(self.context)
    }

    self.context = nil
}