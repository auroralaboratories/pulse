// +build !cgocheck
package pulse

// #include "client.h"
// #cgo pkg-config: libpulse
import "C"

import (
    // log "github.com/Sirupsen/logrus"
)

type Operation struct {
    Client     *Client
    Done       chan error
    Properties map[string]interface{}
    Data       []byte
}


func NewOperation(client *Client) *Operation {
    return &Operation{
        Client:     client,
        Done:       make(chan error),
        Properties: make(map[string]interface{}),
        Data:       make([]byte, 0),
    }
}