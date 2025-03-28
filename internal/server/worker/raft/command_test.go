package raft

import (
	"testing"
)

func TestOpTypeString(t *testing.T) {
	tests := []struct {
		op       OpType
		expected string
	}{
		{OpSet, "set"},
		{OpDelete, "delete"},
		{OpType(999), ""}, // Invalid OpType
	}

	for _, tt := range tests {
		result := tt.op.String()
		if result != tt.expected {
			t.Errorf("OpType.String() = %q, want %q", result, tt.expected)
		}
	}
}

func TestMakeCommand(t *testing.T) {
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
		if result.Op != tt.expected.Op || result.Key != tt.expected.Key || result.Value != tt.expected.Value {
			t.Errorf("makeCommand(%v, %q, %q) = %+v, want %+v", tt.op, tt.key, tt.value, result, tt.expected)
		}
	}
}
