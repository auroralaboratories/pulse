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
        `server-string`:     `test/server:string`,
        `cookie`:            121723128374,
        `nonexistent-field`: false,
        // `channels`:          `wrong-data-type`,
        // `daemon-hostname`:   []string{ `what`, `u`, `say`, `?` },
    }, &v); err != nil {
        t.Errorf("Failed to unmarshal map: %v", err)
    }

    if err := UnmarshalMap(map[string]interface{}{
        `server-string`:     `test/server:string`,
        `cookie`:            121723128374,
        `nonexistent-field`: false,
        `channels`:          `wrong-data-type`,
        `daemon-hostname`:   []string{ `what`, `u`, `say`, `?` },
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

        t.Logf("%+v", my)
    }
}

func TestNewClient(t *testing.T) {
    _, err := NewClient(`test-client-1`)

    if err != nil {
        t.Errorf("%+v", err)
    }
}


func TestGetServerInfo(t *testing.T) {
    if client, err := NewClient(`test-client-1`); err == nil {
        if info, err := client.GetServerInfo(); err != nil {
            t.Errorf("ServerInfo failed: %+v", err)
        }else{
            t.Logf("ServerInfo: %+v", info)
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}

