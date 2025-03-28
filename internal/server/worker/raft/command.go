package raft

type OpType int

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
