package validate

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRaft_Validate(t *testing.T) {
	tests := []struct {
		name          string
		setupViper    func(*viper.Viper)
		expectedError error
	}{
		{
			name: "missing mandatory fields",
			setupViper: func(_ *viper.Viper) {
				// Do not set any required fields
			},
			expectedError: ErrRaftMissingMandatory,
		},
		{
			name: "missing raft-data-dir when devmode is false",
			setupViper: func(v *viper.Viper) {
				v.Set("server.raft.peers", "peer1,peer2")
				v.Set("server.raft.node_id", "node1")
				v.Set("server.raft.devmode", false)
				// Do not set server.raft.datadir
			},
			expectedError: ErrRaftStorage,
		},
		{
			name: "valid configuration with devmode true",
			setupViper: func(v *viper.Viper) {
				v.Set("server.raft.peers", "peer1,peer2")
				v.Set("server.raft.node_id", "node1")
				v.Set("server.raft.devmode", true)
				// server.raft.datadir is not required in devmode
			},
			expectedError: nil,
		},
		{
			name: "valid configuration with devmode false",
			setupViper: func(v *viper.Viper) {
				v.Set("server.raft.peers", "peer1,peer2")
				v.Set("server.raft.node_id", "node1")
				v.Set("server.raft.devmode", false)
				v.Set("server.raft.datadir", "/data/raft")
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viperInstance := viper.New()
			tt.setupViper(viperInstance)

			raft := NewRaft()
			err := raft.Validate(viperInstance)

			assert.ErrorIs(t, tt.expectedError, err)
		})
	}
}
