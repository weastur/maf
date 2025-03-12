package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type Raft struct{}

var ErrRaft = errors.New(
	"raft peers and node ID must be set",
)

func NewRaft() *Raft {
	return &Raft{}
}

func (v *Raft) Validate(viperInstance *viper.Viper) error {
	if !viperInstance.IsSet("server.raft.peers") || !viperInstance.IsSet("server.raft.node_id") {
		return ErrRaft
	}

	return nil
}
