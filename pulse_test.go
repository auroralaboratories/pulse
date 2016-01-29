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

func TestNewContext(t *testing.T) {
    if client, err := NewClient(); err == nil {
        if context, err := client.NewContext(); err != nil {
            t.Errorf("Failed to create context: %+v", err)
        }
    }else{
        t.Errorf("Failed to create client: %+v", err)
    }
}

