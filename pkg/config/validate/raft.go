package validate

import (
	"errors"

	"github.com/spf13/viper"
)

type ValidatorRaft struct{}

var ErrRaft = errors.New(
	"raft peers and node ID must be set",
)

func NewValidatorRaft() *ValidatorRaft {
	return &ValidatorRaft{}
}

func (v *ValidatorRaft) Validate(viperInstance *viper.Viper) error {
	if !viperInstance.IsSet("server.raft.peers") || !viperInstance.IsSet("server.raft.node_id") {
		return ErrRaft
	}

	return nil
}
