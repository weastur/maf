package raft

type Consensus interface {
	IsLeader() bool
}
