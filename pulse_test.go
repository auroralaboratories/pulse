package pulse

import (
    "time"
    "testing"
    // log "github.com/Sirupsen/logrus"
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
            for _, sink := range sinks {
                t.Logf("GetSinks(): %+v", sink)
            }
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}


func TestGetSink0(t *testing.T) {
    if client, err := NewClient(`test-client-get-sink-0`); err == nil {
        if sinks, err := client.GetSinks(); err == nil {
            if len(sinks) > 0 {
                sink := sinks[0]

                if err := sink.Refresh(); err == nil {
                    t.Logf("Sink %d", sink.Index)
                    t.Logf("  Volume: (%f%%) %d / %d", float64(sink.VolumeFactor * 100.0), sink.CurrentVolumeStep, sink.NumVolumeSteps)
                }else{
                    t.Errorf("Failed to refresh sink: %v", err)
                }
            }else{
                t.Errorf("No sinks returned")
            }
        }else{
            t.Errorf("GetSinks() failed: %+v", err)
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}


func TestGetSink0SetVolume(t *testing.T) {
    if client, err := NewClient(`test-client-get-sink-0`); err == nil {
        if sinks, err := client.GetSinks(); err == nil {
            if len(sinks) > 0 {
                sink := sinks[0]

                if err := sink.SetVolume(0.75); err == nil {
                    t.Logf("Sink %d", sink.Index)
                    t.Logf("Volume:    (%f%%) %d / %d", float64(sink.VolumeFactor * 100.0), sink.CurrentVolumeStep, sink.NumVolumeSteps)
                }else{
                    t.Errorf("Failed to set volume: %v", err)
                }

                if err := sink.IncreaseVolume(0.1); err == nil && sink.VolumeFactor == 0.85 {
                    t.Logf("Increased: (%f%%) %d / %d", float64(sink.VolumeFactor * 100.0), sink.CurrentVolumeStep, sink.NumVolumeSteps)
                }else{
                    t.Errorf("Failed to increase volume: %v", err)
                }

                if err := sink.DecreaseVolume(0.1); err == nil && sink.VolumeFactor == 0.75 {
                    t.Logf("Decreased: (%f%%) %d / %d", float64(sink.VolumeFactor * 100.0), sink.CurrentVolumeStep, sink.NumVolumeSteps)
                }else{
                    t.Errorf("Failed to decrease volume: %v", err)
                }
            }else{
                t.Errorf("No sinks returned")
            }
        }else{
            t.Errorf("GetSinks() failed: %+v", err)
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}

func TestGetSink0SetMute(t *testing.T) {
    if client, err := NewClient(`test-client-get-sink-0`); err == nil {
        if sinks, err := client.GetSinks(); err == nil {
            if len(sinks) > 0 {
                sink := sinks[0]

                if err := sink.Mute(); err != nil {
                    t.Errorf("Failed to mute sink: %v", err)
                }else if !sink.Muted {
                    t.Errorf("Failed to mute sink: still not muted")
                }

                time.Sleep(1 * time.Second)

                if err := sink.Unmute(); err != nil {
                    t.Errorf("Failed to unmute sink: %v", err)
                }else if sink.Muted {
                    t.Errorf("Failed to unmute sink: still muted")
                }

                time.Sleep(1 * time.Second)

                if err := sink.ToggleMute(); err != nil {
                    t.Errorf("Failed to toggle mute on: %v", err)
                }else if !sink.Muted {
                    t.Errorf("Failed to toggle mute on: still not muted")
                }

                time.Sleep(1 * time.Second)

                if err := sink.ToggleMute(); err != nil {
                    t.Errorf("Failed to toggle mute off: %v", err)
                }else if sink.Muted {
                    t.Errorf("Failed to toggle mute off: still muted")
                }

            }else{
                t.Errorf("No sinks returned")
            }
        }else{
            t.Errorf("GetSinks() failed: %+v", err)
        }
    }else{
        t.Errorf("Client create failed: %+v", err)
    }
}

