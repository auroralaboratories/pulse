package pulse

type ServerInfo struct {
    Channels               int    `key:"channels"`
    Cookie                 int    `key:"cookie"`
    DaemonHostname         string `key:"daemon-hostname"`
    DaemonUser             string `key:"daemon-user"`
    DefaultSinkName        string `key:"default-sink-name"`
    DefaultSourceName      string `key:"default-source-name"`
    LibraryProtocolVersion int    `key:"library-protocol-version"`
    Name                   string `key:"server-name"`
    ProtocolVersion        int    `key:"server-protocol-version"`
    SampleFormat           string `key:"sample-format"`
    SampleRate             int    `key:"sample-rate"`
    ServerString           string `key:"server-string"`
    Version                string `key:"server-version"`
}

