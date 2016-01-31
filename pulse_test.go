package pulse

import (
    "testing"
)

func TestUnmarshalMap(t *testing.T) {
    v := ServerInfo{}

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

