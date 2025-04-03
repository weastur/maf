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
}

func (m *MockStorage) Get(key string) (string, bool) {
	args := m.Called(key)

	return args.String(0), args.Bool(1)
}

func (m *MockStorage) Set(key, value string) {
	m.Called(key, value)
}

func (m *MockStorage) Delete(key string) {
	m.Called(key)
}

func (m *MockStorage) Snapshot() Mapping {
	args := m.Called()

	return args.Get(0).(Mapping)
}

func (m *MockStorage) Restore(data Mapping) {
	m.Called(data)
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
	t.Parallel()

	storage := &MockStorage{}
	fsm := NewFSM(storage)

	// Test OpSet
	cmd := Command{Op: OpSet, Key: "key1", Value: "value1"}
	mockCall := storage.On("Set", "key1", "value1").Return()
	data, _ := json.Marshal(cmd)
	log := &raft.Log{Data: data}

	fsm.Apply(log)
	storage.AssertExpectations(t)
	mockCall.Unset()

	// Test OpDelete
	cmd = Command{Op: OpDelete, Key: "key1"}
	mockCall = storage.On("Delete", "key1").Return()
	data, _ = json.Marshal(cmd)
	log = &raft.Log{Data: data}

	fsm.Apply(log)
	storage.AssertExpectations(t)
	mockCall.Unset()

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
	t.Parallel()

	storage := &MockStorage{}
	fsm := NewFSM(storage)
	storage.On("Snapshot").Return(Mapping{"key1": "value1"})

	snapshot, err := fsm.Snapshot()
	require.NoError(t, err)

	storage.AssertExpectations(t)

	fsmSnapshot, ok := snapshot.(*FSMSnapshot)
	assert.True(t, ok)
	assert.Equal(t, Mapping{"key1": "value1"}, fsmSnapshot.data)
}

func TestFSM_Restore(t *testing.T) {
	t.Parallel()

	storage := &MockStorage{}
	fsm := NewFSM(storage)
	storage.On("Restore", Mapping{"key1": "value1"}).Return()

	data := map[string]string{"key1": "value1"}
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(data)

	err := fsm.Restore(io.NopCloser(buf))
	require.NoError(t, err)
	storage.AssertExpectations(t)

	// Test invalid JSON
	invalidData := []byte("invalid json")
	buf = bytes.NewBuffer(invalidData)
	err = fsm.Restore(io.NopCloser(buf))
	require.Error(t, err)
}

func TestFSMSnapshot_Persist(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func TestFSMSnapshot_Release(t *testing.T) {
	t.Parallel()

	// No assertions needed, just ensure no panic occurs.
	snapshot := &FSMSnapshot{}
	snapshot.Release()
}
