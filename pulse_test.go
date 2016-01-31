package pulse

import (
    "testing"
)

type MyStruct struct {
    Name  string `key:"name,omitempty"`
    Count int    `key:"count"`
    Other string
}

func TestUnmarshalMap(t *testing.T) {
    v := ServerInfo{}
    my := MyStruct{}

    if err := UnmarshalMap(map[string]interface{}{
        `ServerString`:     `test/server:string`,
        `Cookie`:            121723128374,
        `nonexistent-field`: false,
    }, &v); err != nil {
        t.Errorf("Failed to unmarshal map: %v", err)
    }

    if err := UnmarshalMap(map[string]interface{}{
        `ServerString`:      `test/server:string`,
        `Cookie`:            121723128374,
        `nonexistent-field`: false,
        `Channels`:          `wrong-data-type`,
        `DaemonHostname`:    []string{ `what`, `u`, `say`, `?` },
    }, &v); err == nil {
        t.Errorf("unmarshal map should have failed, but didn't")
    }

    if err := UnmarshalMap(map[string]interface{}{
        `name`:     `TestingName`,
        `count`:    54,
        `Other`:    `Should be here`,
    }, &my); err != nil {
        t.Errorf("Failed to unmarshal map: %v", err)
    }else{
        if my.Name != `TestingName` {
            t.Errorf("MyStruct: expected 'TestingName', got '%s'", my.Name)
        }

        if my.Count != 54 {
            t.Errorf("MyStruct: expected 54, got %d", my.Count)
        }

        if my.Other != `Should be here` {
            t.Errorf("MyStruct: expected 'Should be here', got '%s'", my.Other)
        }
    }
}

func TestNewClient(t *testing.T) {
    _, err := NewClient(`test-client-create`)

    if err != nil {
        t.Errorf("%+v", err)
    }
}


func TestGetServerInfo(t *testing.T) {
    if client, err := NewClient(`test-client-get-server-info`); err == nil {
        if _, err := client.GetServerInfo(); err != nil {
            t.Errorf("GetServerInfo() failed: %+v", err)
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}


func TestGetSinks(t *testing.T) {
    if client, err := NewClient(`test-client-get-sinks`); err == nil {
        if sinks, err := client.GetSinks(); err != nil {
            t.Errorf("GetSinks() failed: %+v", err)
        }else{
            t.Logf("GetSinks(): %+v", sinks)
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}

