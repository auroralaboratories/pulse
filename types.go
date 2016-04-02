package pulse

type ServerInfo struct {
	Channels               int
	Cookie                 int
	DaemonHostname         string
	DaemonUser             string
	DefaultSinkName        string
	DefaultSourceName      string
	LibraryProtocolVersion int
	Name                   string
	ProtocolVersion        int
	SampleFormat           string
	SampleRate             int
	ServerString           string
	Version                string
}
