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

    // log "github.com/Sirupsen/logrus"
)

type Client struct {
    mainloop *C.pa_threaded_mainloop
    api      *C.pa_mainloop_api
}

func NewClient() (*Client, error) {
    rv := &Client{}

    if mainloop := C.pa_threaded_mainloop_new(); mainloop != nil {
        rv.mainloop = mainloop
    }else{
        return nil, fmt.Errorf("Failed to create PulseAudio mainloop")
    }

    return rv, nil
}

func (self *Client) Start() error {
    if self.mainloop != nil {
        if i := int(C.pa_threaded_mainloop_start(self.mainloop)); i < 0 {
            return fmt.Errorf("Failed to start PulseAudio mainloop")
        }

        self.api = C.pa_threaded_mainloop_get_api(self.mainloop)
    }else{
        return fmt.Errorf("Cannot lock non-existent PulseAudio mainloop")
    }

    return nil
}

func (self *Client) Lock() error {
    if self.mainloop != nil {
        C.pa_threaded_mainloop_lock(self.mainloop)
    }else{
        return fmt.Errorf("Cannot lock non-existent PulseAudio mainloop")
    }

    return nil
}

func (self *Client) Unlock() error {
    if self.mainloop != nil {
        C.pa_threaded_mainloop_unlock(self.mainloop)
    }else{
        return fmt.Errorf("Cannot unlock non-existent PulseAudio mainloop")
    }

    return nil
}

func (self *Client) Wait() {
    C.pa_threaded_mainloop_wait(self.mainloop)
}

func (self *Client) Stop() error {
    if self.mainloop != nil {
        defer self.Free()
        C.pa_threaded_mainloop_stop(self.mainloop)
    }else{
        return fmt.Errorf("Cannot stop non-existent PulseAudio mainloop")
    }

    return nil
}

func (self *Client) Free() {
    if self.mainloop != nil {
        C.pa_threaded_mainloop_free(self.mainloop)
    }

    self.mainloop = nil
}

func (self *Client) NewContext(name string, flags ContextFlags) (*Context, error) {
    rv := &Context{
        Client:  self,
        Name:    name,
    }

    if err := rv.Initialize(flags); err == nil {
        return rv, nil
    }else{
        return nil, err
    }
}
