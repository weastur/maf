package validate

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestRaftValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        map[string]any
		expectedError error
	}{
		{
			name:   "missing mandatory fields",
			config: map[string]any{
				// Do not set any required fields
			},
			expectedError: ErrRaftMissingMandatory,
		},
		{
			name: "missing raft-data-dir when devmode is false",
			config: map[string]any{
				"server.raft.peers":   "peer1,peer2",
				"server.raft.node_id": "node1",
				"server.raft.devmode": false,
				// Do not set server.raft.datadir
			},
			expectedError: ErrRaftStorage,
		},
		{
			name: "valid configuration with devmode true",
			config: map[string]any{
				"server.raft.peers":   "peer1,peer2",
				"server.raft.node_id": "node1",
				"server.raft.devmode": true,
				// server.raft.datadir is not required in devmode
			},
			expectedError: nil,
		},
		{
			name: "valid configuration with devmode false",
			config: map[string]any{
				"server.raft.peers":   "peer1,peer2",
				"server.raft.node_id": "node1",
				"server.raft.devmode": false,
				"server.raft.datadir": "/data/raft",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := viper.New()
			for key, value := range tt.config {
				v.Set(key, value)
			}

			raft := NewRaft()
			err := raft.Validate(v)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}
