package v1alpha

type JoinRequest struct {
	ServerID string `json:"serverId" validate:"required"`
	Addr     string `json:"addr"     validate:"required,tcp_addr"`
}

type LeaveRequest struct {
	ServerID string `json:"serverId" validate:"required"`
}
