package raft

type OpType int

const (
	OpSet = iota
	OpDelete
)

func (op OpType) String() string {
	return [...]string{"get", "set", "delete"}[op]
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
