package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weastur/maf/internal/agent/worker/fiber"
)

func TestGet(t *testing.T) {
	config := &Config{
		LogLevel:  "info",
		LogPretty: true,
		SentryDSN: "test-dsn",
	}
	fiberConfig := &fiber.Config{}

	agent1 := Get(config, fiberConfig)
	agent2 := Get(config, fiberConfig)

	assert.Equal(t, agent1, agent2, "Get should return the same instance")
	assert.Equal(t, config, agent1.config, "Agent should have the correct config")
	assert.Equal(t, fiberConfig, agent1.fiberConfig, "Agent should have the correct fiberConfig")
}
