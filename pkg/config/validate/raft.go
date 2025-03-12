package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type Raft struct{}

var ErrRaftMissingMandatory = errors.New(
	"raft peers and node ID must be set",
)

var ErrRaftStorage = errors.New(
	"raft-data-dir must be set when raft-devmode is false",
)

func NewRaft() *Raft {
	return &Raft{}
}

func (v *Raft) Validate(viperInstance *viper.Viper) error {
	if !viperInstance.IsSet("server.raft.peers") || !viperInstance.IsSet("server.raft.node_id") {
		return ErrRaftMissingMandatory
	}

	if !viperInstance.GetBool("server.raft.devmode") && !viperInstance.IsSet("server.raft.data_dir") {
		return ErrRaftStorage
	}

	return nil
}
