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

type raftJoinRequest struct {
	ServerID string `json:"serverId"`
	Addr     string `json:"addr"`
}

type raftKVGetRequest struct {
	Key string `json:"key"`
}

type raftKVSetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type raftKVGetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Exist bool   `json:"exist"`
}

type TLSConfig struct {
	CertFile       string
	KeyFile        string
	ServerCertFile string
}

func (r *response) IsSuccess() bool {
	return r.Status == statusSuccess
}
