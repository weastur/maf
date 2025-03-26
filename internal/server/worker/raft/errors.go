package raft

import "errors"

var ErrNotALeader = errors.New("not a leader")
