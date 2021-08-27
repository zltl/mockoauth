package viewmodule

// ServerInfo is a view module that displays information about the server.
type ServerInfo struct {
	// The host name and port
	// quant67.com:8443 for example
	Host string `json:"host"`
}
