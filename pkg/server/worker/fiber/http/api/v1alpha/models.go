package v1alpha

type JoinRequest struct {
	ServerID string `json:"serverId"`
	Addr     string `json:"addr"`
}

type LeaveRequest struct {
	ServerID string `json:"serverId"`
}
