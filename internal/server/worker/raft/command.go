package raft

import (
	"encoding/json"
	"errors"
)

type OpType int

var ErrInvalidOpType = errors.New("invalid operation type")

const (
	OpSet = iota
	OpDelete
)

func (op OpType) String() string {
	if op < OpSet || op > OpDelete {
		return ""
	}

	return [...]string{"set", "delete"}[op]
}

type Command struct {
	Op    OpType `json:"op"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func makeCommand(op OpType, key, value string) *Command {
	return &Command{
		Op:    op,
		Key:   key,
		Value: value,
	}
}

func (c *Command) MarshalJSON() ([]byte, error) {
	if c.Op < OpSet || c.Op > OpDelete {
		return nil, ErrInvalidOpType
	}

	type Alias Command

	return json.Marshal((*Alias)(c))
}
