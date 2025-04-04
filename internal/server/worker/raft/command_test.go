package raft

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCommandMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		command  *Command
		expected string
		hasError bool
	}{
		{
			command:  &Command{Op: OpSet, Key: "key1", Value: "value1"},
			expected: `{"op":0,"key":"key1","value":"value1"}`,
			hasError: false,
		},
		{
			command:  &Command{Op: OpDelete, Key: "key2"},
			expected: `{"op":1,"key":"key2"}`,
			hasError: false,
		},
		{
			command:  &Command{Op: OpType(999), Key: "key3"},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		result, err := tt.command.MarshalJSON()
		if tt.hasError {
			require.Error(t, err)
			require.ErrorIs(t, err, ErrInvalidOpType)
		} else {
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(result))
		}
	}
}
