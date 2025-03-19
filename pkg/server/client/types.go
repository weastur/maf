package client

const (
	statusSuccess = "success"
	statusError   = "error"
	statusWarning = "warning"
)

type response struct {
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Error  string `json:"error,omitempty"`
}

type joinRequest struct {
	ServerID string `json:"serverId"`
	Addr     string `json:"addr"`
}

func (r *response) IsSuccess() bool {
	return r.Status == statusSuccess
}
