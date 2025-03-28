package raft

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	mock.Mock
	data map[string]string
}

func (m *MockStorage) Get(key string) (string, bool) {
	value, ok := m.data[key]

	return value, ok
}

func (m *MockStorage) Set(key, value string) {
	m.data[key] = value
}

func (m *MockStorage) Delete(key string) {
	delete(m.data, key)
}

func (m *MockStorage) Snapshot() Mapping {
	return Mapping(m.data)
}

func (m *MockStorage) Restore(data Mapping) {
	m.data = map[string]string(data)
}

type MockSnapshotSink struct {
	mock.Mock
}

func (m *MockSnapshotSink) ID() string {
	return m.Called().String(0)
}

func (m *MockSnapshotSink) Write(p []byte) (int, error) {
	args := m.Called(p)

	return args.Int(0), args.Error(1)
}

func (m *MockSnapshotSink) Close() error {
	return m.Called().Error(0)
}

func (m *MockSnapshotSink) Cancel() error {
	return m.Called().Error(0)
}

func TestFSM_Apply(t *testing.T) {
	storage := &MockStorage{data: make(map[string]string)}
	fsm := NewFSM(storage)

	// Test OpSet
	cmd := Command{Op: OpSet, Key: "key1", Value: "value1"}
	data, _ := json.Marshal(cmd)
	log := &raft.Log{Data: data}

	fsm.Apply(log)
	assert.Equal(t, "value1", storage.data["key1"])

	// Test OpDelete
	cmd = Command{Op: OpDelete, Key: "key1"}
	data, _ = json.Marshal(cmd)
	log = &raft.Log{Data: data}

	fsm.Apply(log)

	_, exists := storage.data["key1"]
	assert.False(t, exists)

	// Test unrecognized command
	cmd = Command{Op: 999, Key: "key1"}
	data, _ = json.Marshal(cmd)
	log = &raft.Log{Data: data}

	assert.Panics(t, func() {
		fsm.Apply(log)
	})

	// Test invalid JSON
	data = []byte("invalid json")
	log = &raft.Log{Data: data}

	assert.Panics(t, func() {
		fsm.Apply(log)
	})
}

func TestFSM_Snapshot(t *testing.T) {
	storage := &MockStorage{data: map[string]string{"key1": "value1"}}
	fsm := NewFSM(storage)

	snapshot, err := fsm.Snapshot()
	require.NoError(t, err)

	fsmSnapshot, ok := snapshot.(*FSMSnapshot)
	assert.True(t, ok)
	assert.Equal(t, Mapping{"key1": "value1"}, fsmSnapshot.data)
}

func TestFSM_Restore(t *testing.T) {
	storage := &MockStorage{data: make(map[string]string)}
	fsm := NewFSM(storage)

	data := map[string]string{"key1": "value1"}
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(data)

	err := fsm.Restore(io.NopCloser(buf))
	require.NoError(t, err)
	assert.Equal(t, "value1", storage.data["key1"])

	// Test invalid JSON
	invalidData := []byte("invalid json")
	buf = bytes.NewBuffer(invalidData)
	err = fsm.Restore(io.NopCloser(buf))
	require.Error(t, err)
}

func TestFSMSnapshot_Persist(t *testing.T) {
	data := Mapping{"key1": "value1"}
	snapshot := &FSMSnapshot{data: data}

	mockSink := &MockSnapshotSink{}
	mockSink.On("Write", mock.Anything).Return(len(data), nil)
	mockSink.On("Close").Return(nil)

	err := snapshot.Persist(mockSink)
	require.NoError(t, err)
	mockSink.AssertExpectations(t)
}

func TestFSMSnapshot_Persist_Error(t *testing.T) {
	data := Mapping{"key1": "value1"}
	snapshot := &FSMSnapshot{data: data}

	mockSink := &MockSnapshotSink{}
	mockSink.On("Write", mock.Anything).Return(0, assert.AnError)
	mockSink.On("Cancel").Return(nil)

	err := snapshot.Persist(mockSink)
	require.Error(t, err)
	mockSink.AssertCalled(t, "Cancel")
	mockSink.AssertExpectations(t)

	// Test error during sink.Close
	mockSink = &MockSnapshotSink{}
	mockSink.On("Write", mock.Anything).Return(len(data), nil)
	mockSink.On("Close").Return(assert.AnError)
	mockSink.On("Cancel").Return(nil)

	err = snapshot.Persist(mockSink)
	require.Error(t, err)
	mockSink.AssertCalled(t, "Cancel")
	mockSink.AssertExpectations(t)

	// Test error during sink.Cancel
	mockSink = &MockSnapshotSink{}
	mockSink.On("Write", mock.Anything).Return(len(data), nil)
	mockSink.On("Close").Return(assert.AnError)
	mockSink.On("Cancel").Return(assert.AnError)

	err = snapshot.Persist(mockSink)
	require.Error(t, err)
	mockSink.AssertCalled(t, "Cancel")
	mockSink.AssertExpectations(t)
}

func TestFSMSnapshot_Release(_ *testing.T) {
	// No assertions needed, just ensure no panic occurs.
	snapshot := &FSMSnapshot{}
	snapshot.Release()
}
