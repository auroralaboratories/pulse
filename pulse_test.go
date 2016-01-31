package pulse

import (
    "testing"
)


func TestNewClient(t *testing.T) {
    _, err := NewClient(`test-client-1`)

    if err != nil {
        t.Errorf("%+v", err)
    }
}


// func TestGetSinks(t *testing.T) {
//     if client, err := NewClient(`test-client-1`); err == nil {
//         if err := client.ServerInfo(); err != nil {
//             t.Errorf("ServerInfo failed: %+v", err)
//         }
//     }else{
//         t.Errorf("Client create failed: %+v", err)
//     }
// }
