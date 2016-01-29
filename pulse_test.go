package pulse

import (
    "testing"
)


func TestNewClient(t *testing.T) {
    _, err := NewClient()

    if err != nil {
        t.Errorf("%+v", err)
    }
}

