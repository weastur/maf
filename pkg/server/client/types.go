package client

type response struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

type joinRequest struct {
	ServerID string `json:"serverId"`
	Addr     string `json:"addr"`
}
