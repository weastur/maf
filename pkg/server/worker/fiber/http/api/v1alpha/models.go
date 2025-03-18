package v1alpha

// Join request
// @Description Raft join request with server metadata
type JoinRequest struct {
	ServerID string `example:"maf-2"         json:"serverId" validate:"required"`
	Addr     string `example:"10.1.2.3:7081" json:"addr"     validate:"required,tcp_addr"`
} // @Name JoinRequest

// Leave request
// @Description Raft leave request with server metadata
type LeaveRequest struct {
	ServerID string `example:"maf-2" json:"serverId" validate:"required"`
} // @Name LeaveRequest
