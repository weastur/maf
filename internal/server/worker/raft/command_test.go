package raft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpTypeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op       OpType
		expected string
	}{
		{OpSet, "set"},
		{OpDelete, "delete"},
		{OpType(999), ""}, // Invalid OpType
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.op.String())
	}
}

func TestMakeCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		op       OpType
		key      string
		value    string
		expected *Command
	}{
		{OpSet, "key1", "value1", &Command{Op: OpSet, Key: "key1", Value: "value1"}},
		{OpDelete, "key2", "", &Command{Op: OpDelete, Key: "key2", Value: ""}},
	}

	for _, tt := range tests {
		result := makeCommand(tt.op, tt.key, tt.value)
		assert.Equal(t, tt.expected, result)
	}
}
