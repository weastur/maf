package client

type Response struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

type JoinRequest struct {
	ServerID string `json:"serverId"`
	Addr     string `json:"addr"`
}
