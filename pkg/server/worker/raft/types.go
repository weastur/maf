package raft

type Consensus interface {
	IsLeader() bool
	Join(serverID, addr string) error
	Leave(serverID string) error
}
