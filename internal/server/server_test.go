package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weastur/maf/internal/server/worker/fiber"
	"github.com/weastur/maf/internal/server/worker/raft"
)

func TestGet(t *testing.T) {
	config := &Config{
		LogLevel:  "debug",
		LogPretty: true,
		SentryDSN: "",
	}
	raftConfig := &raft.Config{
		NodeID: "test-node",
	}
	fiberConfig := &fiber.Config{}

	serverInstance := Get(config, raftConfig, fiberConfig)

	assert.NotNil(t, serverInstance)

	assert.Equal(t, config, serverInstance.config)
	assert.Equal(t, raftConfig, serverInstance.raftConfig)
	assert.Equal(t, fiberConfig, serverInstance.fiberConfig)

	secondInstance := Get(nil, nil, nil)

	assert.Equal(t, serverInstance, secondInstance)
}
