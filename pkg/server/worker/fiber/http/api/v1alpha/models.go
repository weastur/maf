package v1alpha

// Join request
// @Description Raft join request with server metadata
type JoinRequest struct {
	ServerID string `example:"maf-2"         json:"serverId" validate:"required"`
	Addr     string `example:"10.1.2.3:7081" json:"addr"     validate:"required,tcp_addr"`
} // @Name RaftJoinRequest

// Leave request
// @Description Raft leave request with server metadata
type LeaveRequest struct {
	ServerID string `example:"maf-2" json:"serverId" validate:"required"`
} // @Name RaftLeaveRequest

// Server metadata
// @Description Metadata of the server in the raft cluster
type Server struct {
	ID      string `example:"maf-1"          json:"id"`
	Address string `example:"127.0.0.1:7081" json:"address"`
	// Suffrage of the server in terms of the consensus: Voter, Nonvoter, Staging
	Suffrage string `enums:"Voter,Nonvoter,Staging" example:"Voter" json:"suffrage"`
	Leader   bool   `example:"true"                 json:"leader"`
} // @Name RaftServer

// Info response
// @Description Satatus of the raft cluster with servers metadata
type InfoResponse struct {
	// State of the server in terms of the consensus: Leader, Follower, Candidate, etc.
	State string `enums:"Follower,Candidate,Leader,Shutdown,Unknown" example:"Leader" json:"state"`
	Addr  string `example:"127.0.0.1:7081"                           json:"addr"`
	ID    string `example:"maf-1"                                    json:"id"`
	// List of servers in the cluster
	Servers []Server `json:"servers"`
	// Extended stats of the raft cluster
	Stats map[string]string `json:"stats"`
} // @Name RaftInfoResponse

// KV get response
// @Description Response to the get request.
// @Description Also contains 'exist' flag to distinguish between empty and non-existent string value
type KVGetResponse struct {
	Key   string `example:"key"   json:"key"`
	Value string `example:"value" json:"value"`
	Exist bool   `example:"true"  json:"exist"`
} // @Name KVGetResponse

// KV set request
// @Description Request to set the key-value pair
type KVSetRequest struct {
	Key   string `example:"key"   json:"key"`
	Value string `example:"value" json:"value"`
} // @Name KVSetRequest
