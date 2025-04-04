package raft

import (
	"errors"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	hraft "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	hclogzerolog "github.com/weastur/hclog-zerolog"
	"github.com/weastur/maf/internal/utils/logging"
	sentryWrapper "github.com/weastur/maf/internal/utils/sentry"
)

const invalidDir = "/invalid/path"

type MockSentry struct {
	mock.Mock
}

func (m *MockSentry) Flush() {
	m.Called()
}

func (m *MockSentry) Recover() {
	m.Called()
}

func (m *MockSentry) IsConfigured() bool {
	args := m.Called()

	return args.Bool(0)
}

func (m *MockSentry) GetHub() *sentry.Hub {
	args := m.Called()

	return args.Get(0).(*sentry.Hub)
}

func (m *MockSentry) Fork(scopeTag string) *sentryWrapper.Wrapper {
	args := m.Called(scopeTag)

	return args.Get(0).(*sentryWrapper.Wrapper)
}

type MockHRaft struct {
	mock.Mock
}

func (m *MockHRaft) State() hraft.RaftState {
	args := m.Called()

	return args.Get(0).(hraft.RaftState)
}

func (m *MockHRaft) BootstrapCluster(configuration hraft.Configuration) hraft.Future {
	args := m.Called(configuration)

	return args.Get(0).(hraft.Future)
}

func (m *MockHRaft) LeadershipTransfer() hraft.Future {
	args := m.Called()

	return args.Get(0).(hraft.Future)
}

func (m *MockHRaft) Shutdown() hraft.Future {
	args := m.Called()

	return args.Get(0).(hraft.Future)
}

func (m *MockHRaft) GetConfiguration() hraft.ConfigurationFuture {
	args := m.Called()

	return args.Get(0).(hraft.ConfigurationFuture)
}

func (m *MockHRaft) RemoveServer(id hraft.ServerID, prevIndex uint64, timeout time.Duration) hraft.IndexFuture {
	args := m.Called(id, prevIndex, timeout)

	return args.Get(0).(hraft.IndexFuture)
}

func (m *MockHRaft) AddVoter(
	id hraft.ServerID,
	address hraft.ServerAddress,
	prevIndex uint64,
	timeout time.Duration,
) hraft.IndexFuture {
	args := m.Called(id, address, prevIndex, timeout)

	return args.Get(0).(hraft.IndexFuture)
}

func (m *MockHRaft) LeaderWithID() (hraft.ServerAddress, hraft.ServerID) {
	args := m.Called()

	return args.Get(0).(hraft.ServerAddress), args.Get(1).(hraft.ServerID)
}

func (m *MockHRaft) Stats() map[string]string {
	args := m.Called()

	return args.Get(0).(map[string]string)
}

func (m *MockHRaft) Apply(cmd []byte, timeout time.Duration) hraft.ApplyFuture {
	args := m.Called(cmd, timeout)

	return args.Get(0).(hraft.ApplyFuture)
}

func (m *MockHRaft) LeaderCh() <-chan bool {
	args := m.Called()

	return args.Get(0).(<-chan bool)
}

type MockFSM struct {
	mock.Mock
}

func (m *MockFSM) Apply(log *hraft.Log) interface{} {
	args := m.Called(log)

	return args.Get(0)
}

func (m *MockFSM) Snapshot() (hraft.FSMSnapshot, error) {
	args := m.Called()

	return args.Get(0).(hraft.FSMSnapshot), args.Error(1)
}

func (m *MockFSM) Restore(rc io.ReadCloser) error {
	args := m.Called(rc)

	return args.Error(0)
}

type MockLogStore struct {
	mock.Mock
}

func (m *MockLogStore) FirstIndex() (uint64, error) {
	args := m.Called()

	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockLogStore) LastIndex() (uint64, error) {
	args := m.Called()

	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockLogStore) GetLog(index uint64, log *hraft.Log) error {
	args := m.Called(index, log)

	return args.Error(0)
}

func (m *MockLogStore) StoreLog(log *hraft.Log) error {
	args := m.Called(log)

	return args.Error(0)
}

func (m *MockLogStore) StoreLogs(logs []*hraft.Log) error {
	args := m.Called(logs)

	return args.Error(0)
}

func (m *MockLogStore) DeleteRange(min, max uint64) error { //nolint:predeclared,revive
	args := m.Called(min, max)

	return args.Error(0)
}

type MockStableStore struct {
	mock.Mock
}

func (m *MockStableStore) Set(key []byte, val []byte) error {
	args := m.Called(key, val)

	return args.Error(0)
}

func (m *MockStableStore) Get(key []byte) ([]byte, error) {
	args := m.Called(key)

	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStableStore) SetUint64(key []byte, val uint64) error {
	args := m.Called(key, val)

	return args.Error(0)
}

func (m *MockStableStore) GetUint64(key []byte) (uint64, error) {
	args := m.Called(key)

	return args.Get(0).(uint64), args.Error(1)
}

type MockSnapshotStore struct {
	mock.Mock
}

func (m *MockSnapshotStore) Create(
	version hraft.SnapshotVersion,
	index, term uint64,
	configuration hraft.Configuration,
	configurationIndex uint64,
	trans hraft.Transport,
) (hraft.SnapshotSink, error) {
	args := m.Called(version, index, term, configuration, configurationIndex, trans)

	return args.Get(0).(hraft.SnapshotSink), args.Error(1)
}

func (m *MockSnapshotStore) List() ([]*hraft.SnapshotMeta, error) {
	args := m.Called()

	return args.Get(0).([]*hraft.SnapshotMeta), args.Error(1)
}

func (m *MockSnapshotStore) Open(id string) (*hraft.SnapshotMeta, io.ReadCloser, error) {
	args := m.Called(id)

	return args.Get(0).(*hraft.SnapshotMeta), args.Get(1).(io.ReadCloser), args.Error(2)
}

type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Consumer() <-chan hraft.RPC {
	args := m.Called()

	return args.Get(0).(<-chan hraft.RPC)
}

func (m *MockTransport) LocalAddr() hraft.ServerAddress {
	args := m.Called()

	return args.Get(0).(hraft.ServerAddress)
}

func (m *MockTransport) AppendEntriesPipeline(
	id hraft.ServerID,
	target hraft.ServerAddress,
) (hraft.AppendPipeline, error) {
	args := m.Called(id, target)

	return args.Get(0).(hraft.AppendPipeline), args.Error(1)
}

func (m *MockTransport) AppendEntries(
	id hraft.ServerID,
	target hraft.ServerAddress,
	args *hraft.AppendEntriesRequest,
	resp *hraft.AppendEntriesResponse,
) error {
	callArgs := m.Called(id, target, args, resp)

	return callArgs.Error(0)
}

func (m *MockTransport) RequestVote(
	id hraft.ServerID,
	target hraft.ServerAddress,
	args *hraft.RequestVoteRequest,
	resp *hraft.RequestVoteResponse,
) error {
	callArgs := m.Called(id, target, args, resp)

	return callArgs.Error(0)
}

func (m *MockTransport) InstallSnapshot(
	id hraft.ServerID,
	target hraft.ServerAddress,
	args *hraft.InstallSnapshotRequest,
	resp *hraft.InstallSnapshotResponse,
	data io.Reader,
) error {
	callArgs := m.Called(id, target, args, resp, data)

	return callArgs.Error(0)
}

func (m *MockTransport) EncodePeer(id hraft.ServerID, addr hraft.ServerAddress) []byte {
	args := m.Called(id, addr)

	return args.Get(0).([]byte)
}

func (m *MockTransport) DecodePeer(data []byte) hraft.ServerAddress {
	args := m.Called(data)

	return args.Get(0).(hraft.ServerAddress)
}

func (m *MockTransport) SetHeartbeatHandler(cb func(rpc hraft.RPC)) {
	m.Called(cb)
}

func (m *MockTransport) TimeoutNow(
	id hraft.ServerID,
	target hraft.ServerAddress,
	args *hraft.TimeoutNowRequest,
	resp *hraft.TimeoutNowResponse,
) error {
	callArgs := m.Called(id, target, args, resp)

	return callArgs.Error(0)
}

type MockApplyFuture struct {
	mock.Mock
}

func (m *MockApplyFuture) Error() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockApplyFuture) Index() uint64 {
	args := m.Called()

	return args.Get(0).(uint64)
}

func (m *MockApplyFuture) Response() any {
	args := m.Called()

	return args.Get(0)
}

type MockConfigurationFuture struct {
	mock.Mock
}

func (m *MockConfigurationFuture) Error() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockConfigurationFuture) Index() uint64 {
	args := m.Called()

	return args.Get(0).(uint64)
}

func (m *MockConfigurationFuture) Configuration() hraft.Configuration {
	args := m.Called()

	return args.Get(0).(hraft.Configuration)
}

type MockFuture struct {
	mock.Mock
}

func (m *MockFuture) Error() error {
	args := m.Called()

	return args.Error(0)
}

type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) RaftJoin(nodeID, addr string) error {
	args := m.Called(nodeID, addr)

	return args.Error(0)
}

func (m *MockAPIClient) Close() error {
	args := m.Called()

	return args.Error(0)
}

type MockIndexFuture struct {
	mock.Mock
}

func (m *MockIndexFuture) Error() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockIndexFuture) Index() uint64 {
	args := m.Called()

	return args.Get(0).(uint64)
}

func mockTransportWithAddr(addr string) hraft.Transport {
	mockTransport := new(MockTransport)
	mockTransport.On("LocalAddr").Return(hraft.ServerAddress(addr))

	return mockTransport
}

func TestNew(t *testing.T) {
	t.Parallel()

	config := &Config{
		Addr:      "127.0.0.1:8080",
		NodeID:    "node1",
		Devmode:   true,
		Peers:     []string{"http://127.0.0.1:8081"},
		Datadir:   "/tmp/raft",
		Bootstrap: true,
	}
	mockSentry := new(MockSentry)

	raft := New(config, mockSentry)

	assert.Equal(t, config, raft.config, "expected config to match")
	assert.Equal(t, mockSentry, raft.sentry, "expected sentry to match")
	assert.NotNil(t, raft.done, "expected done channel to be initialized")
	assert.Empty(t, raft.leadershipChangesChannels, "expected leadershipChangesChannels to be empty")
}

func TestIsReady(t *testing.T) {
	t.Parallel()

	t.Run("NotReady", func(t *testing.T) {
		t.Parallel()

		raft := &Raft{
			initCompleted: atomic.Bool{},
		}

		assert.False(t, raft.IsReady(), "expected IsReady to return false")
	})

	t.Run("Ready", func(t *testing.T) {
		t.Parallel()

		raft := &Raft{
			initCompleted: atomic.Bool{},
		}
		raft.initCompleted.Store(true)

		assert.True(t, raft.IsReady(), "expected IsReady to return true")
	})
}

func TestIsLive(t *testing.T) {
	t.Parallel()

	t.Run("LeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			raftInstance: mockRaft,
		}

		assert.True(t, raft.IsLive(), "expected IsLive to return true for Leader state")
		mockRaft.AssertExpectations(t)
	})

	t.Run("FollowerState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Follower)

		raft := &Raft{
			raftInstance: mockRaft,
		}

		assert.True(t, raft.IsLive(), "expected IsLive to return true for Follower state")
		mockRaft.AssertExpectations(t)
	})

	t.Run("OtherState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Candidate)

		raft := &Raft{
			raftInstance: mockRaft,
		}

		assert.False(t, raft.IsLive(), "expected IsLive to return false for non-Leader and non-Follower state")
		mockRaft.AssertExpectations(t)
	})
}

func TestEnsureDatadir(t *testing.T) {
	t.Parallel()

	t.Run("ValidDatadir", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		config := &Config{
			Datadir: tempDir,
		}
		raft := &Raft{
			config: config,
		}

		assert.NotPanics(t, func() {
			raft.ensureDatadir()
		}, "expected ensureDatadir to not panic for valid directory")

		info, err := os.Stat(tempDir)
		require.NoError(t, err, "expected no error when checking directory")
		assert.True(t, info.IsDir(), "expected path to be a directory")
	})

	t.Run("EmptyDatadir", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Datadir: "",
		}
		raft := &Raft{
			config: config,
		}

		assert.NotPanics(t, func() {
			raft.ensureDatadir()
		}, "expected ensureDatadir to not panic for empty directory")
	})

	t.Run("InvalidDatadir", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Datadir: invalidDir,
		}
		raft := &Raft{
			config: config,
		}

		assert.Panics(t, func() {
			raft.ensureDatadir()
		}, "expected ensureDatadir to panic for invalid directory")
	})
}

func TestConfigureRaft(t *testing.T) {
	t.Parallel()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			NodeID: "node1",
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.NotPanics(t, func() {
			raft.configureRaft()
		}, "expected configureRaft to not panic")

		assert.NotNil(t, raft.hrconfig, "expected hrconfig to be initialized")
		assert.Equal(t, hraft.ServerID(config.NodeID), raft.hrconfig.LocalID, "expected LocalID to match NodeID")
		assert.Equal(t, raft.hlogger, raft.hrconfig.Logger, "expected Logger to match hlogger")
	})
}

func TestInitTransport(t *testing.T) {
	t.Parallel()

	t.Run("ValidAddress", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Addr: "127.0.0.1:8080",
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.NotPanics(t, func() {
			raft.initTransport()
		}, "expected initTransport to not panic for valid address")

		assert.NotNil(t, raft.transport, "expected transport to be initialized")
		assert.Equal(t, config.Addr, string(raft.transport.LocalAddr()), "expected transport address to match config address")
	})

	t.Run("InvalidAddress", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Addr: "invalid-address",
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.Panics(t, func() {
			raft.initTransport()
		}, "expected initTransport to panic for invalid address")
	})
}

func TestInitSnapshotStore(t *testing.T) {
	t.Parallel()

	t.Run("ValidDatadir", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		config := &Config{
			Datadir: tempDir,
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.NotPanics(t, func() {
			raft.initSnapshotStore()
		}, "expected initSnapshotStore to not panic for valid directory")

		assert.NotNil(t, raft.snapshotStore, "expected snapshotStore to be initialized")
	})

	t.Run("InvalidDatadir", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Datadir: invalidDir,
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.Panics(t, func() {
			raft.initSnapshotStore()
		}, "expected initSnapshotStore to panic for invalid directory")
	})
}

func TestInitStore(t *testing.T) {
	t.Parallel()

	t.Run("DevmodeEnabled", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Devmode: true,
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.NotPanics(t, func() {
			raft.initStore()
		}, "expected initStore to not panic when Devmode is enabled")

		assert.NotNil(t, raft.logStore, "expected logStore to be initialized")
		assert.NotNil(t, raft.stableStore, "expected stableStore to be initialized")
		_, isInmemStore := raft.logStore.(*hraft.InmemStore)
		assert.True(t, isInmemStore, "expected logStore to be an in-memory store")
		_, isInmemStore = raft.stableStore.(*hraft.InmemStore)
		assert.True(t, isInmemStore, "expected stableStore to be an in-memory store")
	})

	t.Run("DevmodeDisabledWithValidDatadir", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		config := &Config{
			Devmode: false,
			Datadir: tempDir,
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.NotPanics(t, func() {
			raft.initStore()
		}, "expected initStore to not panic when Devmode is disabled and Datadir is valid")

		assert.NotNil(t, raft.logStore, "expected logStore to be initialized")
		assert.NotNil(t, raft.stableStore, "expected stableStore to be initialized")
		_, isBoltDB := raft.logStore.(*raftboltdb.BoltStore)
		assert.True(t, isBoltDB, "expected logStore to be a BoltDB store")
		_, isBoltDB = raft.stableStore.(*raftboltdb.BoltStore)
		assert.True(t, isBoltDB, "expected stableStore to be a BoltDB store")
	})

	t.Run("DevmodeDisabledWithInvalidDatadir", func(t *testing.T) {
		t.Parallel()

		config := &Config{
			Devmode: false,
			Datadir: invalidDir,
		}
		raft := &Raft{
			config:  config,
			hlogger: hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		}

		assert.Panics(t, func() {
			raft.initStore()
		}, "expected initStore to panic when Devmode is disabled and Datadir is invalid")
	})
}

func TestInitFSM(t *testing.T) {
	t.Parallel()

	t.Run("FSMInitialization", func(t *testing.T) {
		t.Parallel()

		raft := &Raft{}
		assert.Nil(t, raft.fsm, "expected fsm to be nil before initialization")
		assert.Nil(t, raft.storage, "expected storage to be nil before initialization")

		assert.NotPanics(t, func() {
			raft.initFSM()
		}, "expected initFSM to not panic")

		assert.NotNil(t, raft.fsm, "expected fsm to be initialized")
		assert.NotNil(t, raft.storage, "expected storage to be initialized")
	})
}

func TestInitRaftInstance(t *testing.T) {
	t.Parallel()

	t.Run("InitializationFailure", func(t *testing.T) {
		t.Parallel()

		raft := &Raft{
			hrconfig:      nil, // Invalid configuration
			fsm:           nil,
			logStore:      nil,
			stableStore:   nil,
			snapshotStore: nil,
			transport:     nil,
		}

		assert.Panics(t, func() {
			raft.initRaftInstance()
		}, "expected initRaftInstance to panic with invalid configuration")
	})
}

func TestMonitorLeadership(t *testing.T) {
	t.Parallel()

	t.Run("LeadershipChangeToLeader", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		leaderCh := make(chan bool, 1)

		var recvLeaderCh <-chan bool = leaderCh

		mockRaft.On("LeaderCh").Return(recvLeaderCh)

		raft := &Raft{
			raftInstance: mockRaft,
			done:         make(chan struct{}),
		}

		leadershipChanges := make(LeadershipChangesCh, 1)
		raft.SubscribeOnLeadershipChanges(leadershipChanges)

		go raft.monitorLeadership()

		leaderCh <- true

		select {
		case isLeader := <-leadershipChanges:
			assert.True(t, isLeader, "expected leadership change to leader")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for leadership change")
		}

		close(raft.done)
		mockRaft.AssertExpectations(t)
	})

	t.Run("LeadershipChangeToFollower", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		leaderCh := make(chan bool, 1)

		var recvLeaderCh <-chan bool = leaderCh

		mockRaft.On("LeaderCh").Return(recvLeaderCh)

		raft := &Raft{
			raftInstance: mockRaft,
			done:         make(chan struct{}),
		}

		leadershipChanges := make(LeadershipChangesCh, 1)
		raft.SubscribeOnLeadershipChanges(leadershipChanges)

		go raft.monitorLeadership()

		leaderCh <- false

		select {
		case isLeader := <-leadershipChanges:
			assert.False(t, isLeader, "expected leadership change to follower")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for leadership change")
		}

		close(raft.done)
		mockRaft.AssertExpectations(t)
	})

	t.Run("ChannelOverflow", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		leaderCh := make(chan bool, 1)

		var recvLeaderCh <-chan bool = leaderCh

		mockRaft.On("LeaderCh").Return(recvLeaderCh)

		raft := &Raft{
			raftInstance: mockRaft,
			done:         make(chan struct{}),
		}

		leadershipChanges := make(LeadershipChangesCh, 1)
		raft.SubscribeOnLeadershipChanges(leadershipChanges)

		go raft.monitorLeadership()

		// Simulate a burst of leadership changes
		// This shouldn't block the raft and the leaderCh
		leaderCh <- false
		leaderCh <- true
		leaderCh <- false

		<-time.After(100 * time.Millisecond)

		close(raft.done)
		mockRaft.AssertExpectations(t)
	})

	t.Run("StopMonitoring", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		leaderCh := make(chan bool)

		var recvLeaderCh <-chan bool = leaderCh

		mockRaft.On("LeaderCh").Return(recvLeaderCh)

		raft := &Raft{
			raftInstance: mockRaft,
			done:         make(chan struct{}),
		}

		go raft.monitorLeadership()

		close(raft.done)

		<-time.After(300 * time.Millisecond)

		select {
		case leaderCh <- true:
			t.Fatal("expected monitoring to stop, but channel is still active")
		case <-time.After(100 * time.Millisecond):
		}

		mockRaft.AssertExpectations(t)
	})
}

func TestIsLeader(t *testing.T) {
	t.Parallel()

	t.Run("LeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			raftInstance: mockRaft,
		}

		assert.True(t, raft.IsLeader(), "expected IsLeader to return true for Leader state")
		mockRaft.AssertExpectations(t)
	})

	t.Run("NonLeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Follower)

		raft := &Raft{
			raftInstance: mockRaft,
		}

		assert.False(t, raft.IsLeader(), "expected IsLeader to return false for non-Leader state")
		mockRaft.AssertExpectations(t)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("KeyExists", func(t *testing.T) {
		t.Parallel()

		mockStorage := new(MockStorage)
		mockStorage.On("Get", "existingKey").Return("value", true)

		raft := &Raft{
			storage: mockStorage,
		}

		value, found := raft.Get("existingKey")
		assert.True(t, found, "expected key to be found")
		assert.Equal(t, "value", value, "expected value to match")
		mockStorage.AssertExpectations(t)
	})

	t.Run("KeyDoesNotExist", func(t *testing.T) {
		t.Parallel()

		mockStorage := new(MockStorage)
		mockStorage.On("Get", "missingKey").Return("", false)

		raft := &Raft{
			storage: mockStorage,
		}

		value, found := raft.Get("missingKey")
		assert.False(t, found, "expected key to not be found")
		assert.Empty(t, value, "expected value to be empty")
		mockStorage.AssertExpectations(t)
	})
}

func TestApplyCommand(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulApply", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)
		mockApplyFuture.On("Error").Return(nil)

		mockRaft.On("Apply", mock.Anything, cmdTimeout).Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.applyCommand(OpSet, "key", "value")
		require.NoError(t, err, "expected applyCommand to succeed")
		mockRaft.AssertExpectations(t)
		mockApplyFuture.AssertExpectations(t)
	})

	t.Run("MarshalError", func(t *testing.T) {
		t.Parallel()

		raft := &Raft{
			logger: log.Logger,
		}

		// Intentionally passing an invalid operation type to cause a marshal error
		err := raft.applyCommand(OpType(999), "key", "value")
		require.ErrorIs(t, err, ErrInvalidOpType, "expected applyCommand to fail with ErrInvalidOpType")
	})

	t.Run("ApplyError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)
		mockApplyFuture.On("Error").Return(errors.New("apply error"))

		mockRaft.On("Apply", mock.Anything, cmdTimeout).Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.applyCommand(OpSet, "key", "value")
		require.Error(t, err, "expected applyCommand to fail due to apply error")
		mockRaft.AssertExpectations(t)
		mockApplyFuture.AssertExpectations(t)
	})
}

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("LeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)
		mockApplyFuture.On("Error").Return(nil)

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("Apply", mock.Anything, cmdTimeout).Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Set("key", "value")
		require.NoError(t, err, "expected Set to succeed for leader state")
		mockRaft.AssertExpectations(t)
		mockApplyFuture.AssertExpectations(t)
	})

	t.Run("NonLeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Follower)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Set("key", "value")
		require.NoError(t, err, "expected Set to not return an error for non-leader state")
		mockRaft.AssertExpectations(t)
	})

	t.Run("ApplyError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)
		mockApplyFuture.On("Error").Return(errors.New("apply error"))

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("Apply", mock.Anything, cmdTimeout).Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Set("key", "value")
		require.Error(t, err, "expected Set to fail due to apply error")
		mockRaft.AssertExpectations(t)
		mockApplyFuture.AssertExpectations(t)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("LeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)
		mockApplyFuture.On("Error").Return(nil)

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("Apply", mock.Anything, cmdTimeout).Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Delete("key")
		require.NoError(t, err, "expected Delete to succeed for leader state")
		mockRaft.AssertExpectations(t)
		mockApplyFuture.AssertExpectations(t)
	})

	t.Run("NonLeaderState", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Follower)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Delete("key")
		require.NoError(t, err, "expected Delete to not return an error for non-leader state")
		mockRaft.AssertExpectations(t)
	})

	t.Run("ApplyError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)
		mockApplyFuture.On("Error").Return(errors.New("apply error"))

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("Apply", mock.Anything, cmdTimeout).Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Delete("key")
		require.Error(t, err, "expected Delete to fail due to apply error")
		mockRaft.AssertExpectations(t)
		mockApplyFuture.AssertExpectations(t)
	})
}

func TestGetInfo(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulGetInfo", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:       "node1",
					Address:  "127.0.0.1:8080",
					Suffrage: hraft.Voter,
				},
			},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("LeaderWithID").Return(hraft.ServerAddress("127.0.0.1:8080"), hraft.ServerID("node1"))
		mockRaft.On("Stats").Return(map[string]string{"key": "value"})

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
		}

		info, err := raft.GetInfo(true)
		require.NoError(t, err, "expected GetInfo to succeed")
		assert.Equal(t, "node1", info.ID, "expected node ID to match")
		assert.Equal(t, "127.0.0.1:8080", info.Addr, "expected address to match")
		assert.Equal(t, "Leader", info.State, "expected state to match")
		assert.Len(t, info.Servers, 1, "expected one server in the cluster")
		assert.Equal(t, "node1", info.Servers[0].ID, "expected server ID to match")
		assert.Equal(t, "127.0.0.1:8080", info.Servers[0].Address, "expected server address to match")
		assert.True(t, info.Servers[0].Leader, "expected server to be the leader")
		assert.Equal(t, Stats{"key": "value"}, info.Stats, "expected stats to match")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
	})

	t.Run("GetConfigurationError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockConfigFuture.On("Error").Return(errors.New("configuration error"))
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)

		raft := &Raft{
			raftInstance: mockRaft,
		}

		info, err := raft.GetInfo(false)
		require.Error(t, err, "expected GetInfo to fail due to configuration error")
		assert.Nil(t, info, "expected info to be nil")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
	})

	t.Run("VerboseFalse", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:       "node1",
					Address:  "127.0.0.1:8080",
					Suffrage: hraft.Voter,
				},
			},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("LeaderWithID").Return(hraft.ServerAddress("127.0.0.1:8080"), hraft.ServerID("node1"))

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
		}

		info, err := raft.GetInfo(false)
		require.NoError(t, err, "expected GetInfo to succeed")
		assert.Equal(t, "node1", info.ID, "expected node ID to match")
		assert.Equal(t, "127.0.0.1:8080", info.Addr, "expected address to match")
		assert.Equal(t, "Leader", info.State, "expected state to match")
		assert.Len(t, info.Servers, 1, "expected one server in the cluster")
		assert.Equal(t, "node1", info.Servers[0].ID, "expected server ID to match")
		assert.Equal(t, "127.0.0.1:8080", info.Servers[0].Address, "expected server address to match")
		assert.True(t, info.Servers[0].Leader, "expected server to be the leader")
		assert.Nil(t, info.Stats, "expected stats to be nil when verbose is false")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
	})
}

func TestStop(t *testing.T) {
	t.Parallel()

	t.Run("NotInitialized", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
			done:         make(chan struct{}),
		}

		raft.initCompleted.Store(false)

		assert.NotPanics(t, func() {
			raft.Stop()
		}, "expected Stop to not panic when raft is not initialized")

		mockRaft.AssertNotCalled(t, "Shutdown")
		mockRaft.AssertNotCalled(t, "LeadershipTransfer")
	})

	t.Run("InitializedNotLeader", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockFuture := new(MockFuture)

		mockRaft.On("State").Return(hraft.Follower)
		mockRaft.On("Shutdown").Return(mockFuture)
		mockFuture.On("Error").Return(nil)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
			done:         make(chan struct{}),
		}

		raft.initCompleted.Store(true)

		assert.NotPanics(t, func() {
			raft.Stop()
		}, "expected Stop to not panic when raft is initialized and not a leader")

		mockRaft.AssertCalled(t, "Shutdown")
		mockRaft.AssertNotCalled(t, "LeadershipTransfer")
	})

	t.Run("InitializedLeader", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockFuture := new(MockFuture)
		mockApplyFuture := new(MockApplyFuture)

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("LeadershipTransfer").Return(mockApplyFuture)
		mockRaft.On("Shutdown").Return(mockFuture)
		mockFuture.On("Error").Return(nil)
		mockApplyFuture.On("Error").Return(nil)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
			done:         make(chan struct{}),
		}

		raft.initCompleted.Store(true)

		assert.NotPanics(t, func() {
			raft.Stop()
		}, "expected Stop to not panic when raft is initialized and is a leader")

		mockRaft.AssertCalled(t, "LeadershipTransfer")
		mockRaft.AssertCalled(t, "Shutdown")
	})

	t.Run("ShutdownError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)

		mockRaft.On("State").Return(hraft.Follower)

		mockApplyFuture := new(MockApplyFuture)

		mockApplyFuture.On("Error").Return(errors.New("shutdown error"))
		mockRaft.On("Shutdown").Return(mockApplyFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
			done:         make(chan struct{}),
		}

		raft.initCompleted.Store(true)

		assert.NotPanics(t, func() {
			raft.Stop()
		}, "expected Stop to not panic even if Shutdown returns an error")

		mockRaft.AssertCalled(t, "Shutdown")
		mockApplyFuture.AssertExpectations(t)
	})

	t.Run("LeadershipTransferError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockApplyFuture := new(MockApplyFuture)

		mockRaft.On("State").Return(hraft.Leader)

		mockTransferFuture := new(MockApplyFuture)

		mockTransferFuture.On("Error").Return(errors.New("transfer error"))
		mockRaft.On("LeadershipTransfer").Return(mockTransferFuture)
		mockRaft.On("Shutdown").Return(mockApplyFuture)
		mockApplyFuture.On("Error").Return(nil)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
			done:         make(chan struct{}),
		}

		raft.initCompleted.Store(true)

		assert.NotPanics(t, func() {
			raft.Stop()
		}, "expected Stop to not panic even if LeadershipTransfer returns an error")

		mockRaft.AssertCalled(t, "LeadershipTransfer")
		mockRaft.AssertCalled(t, "Shutdown")
		mockTransferFuture.AssertExpectations(t)
	})
}

func TestBootstrap(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulBootstrap", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockFuture := new(MockFuture)
		mockFuture.On("Error").Return(nil)
		mockRaft.On("BootstrapCluster", mock.Anything).Return(mockFuture)

		raft := &Raft{
			hrconfig: &hraft.Config{
				LocalID: hraft.ServerID("node1"),
			},
			transport:    mockTransportWithAddr("127.0.0.1:8080"),
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		assert.NotPanics(t, func() {
			raft.bootstrap()
		}, "expected bootstrap to not panic")

		assert.True(t, raft.initCompleted.Load(), "expected initCompleted to be set to true")
		mockRaft.AssertExpectations(t)
		mockFuture.AssertExpectations(t)
	})

	t.Run("ClusterAlreadyExists", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockFuture := new(MockFuture)
		mockFuture.On("Error").Return(hraft.ErrCantBootstrap)
		mockRaft.On("BootstrapCluster", mock.Anything).Return(mockFuture)

		raft := &Raft{
			hrconfig: &hraft.Config{
				LocalID: hraft.ServerID("node1"),
			},
			transport:    mockTransportWithAddr("127.0.0.1:8080"),
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		assert.NotPanics(t, func() {
			raft.bootstrap()
		}, "expected bootstrap to not panic when cluster already exists")

		assert.True(t, raft.initCompleted.Load(), "expected initCompleted to be set to true")
		mockRaft.AssertExpectations(t)
		mockFuture.AssertExpectations(t)
	})

	t.Run("BootstrapError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockFuture := new(MockFuture)
		mockFuture.On("Error").Return(errors.New("bootstrap error"))
		mockRaft.On("BootstrapCluster", mock.Anything).Return(mockFuture)

		raft := &Raft{
			hrconfig: &hraft.Config{
				LocalID: hraft.ServerID("node1"),
			},
			transport:    mockTransportWithAddr("127.0.0.1:8080"),
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		assert.Panics(t, func() {
			raft.bootstrap()
		}, "expected bootstrap to panic on error")

		assert.False(t, raft.initCompleted.Load(), "expected initCompleted to remain false")
		mockRaft.AssertExpectations(t)
		mockFuture.AssertExpectations(t)
	})
}

func TestRetryJoin(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulJoin", func(t *testing.T) {
		t.Parallel()

		mockAPIClient := new(MockAPIClient)
		mockAPIClient.On("RaftJoin", "node1", "127.0.0.1:8080").Return(nil)
		mockAPIClient.On("Close").Return(nil)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
				Peers:  []string{"http://127.0.0.1:8081"},
			},
			logger: log.Logger,
			getAPIClient: func(_ string) APIClient {
				return mockAPIClient
			},
			initCompleted: atomic.Bool{},
		}

		go raft.retryJoin()

		time.Sleep(100 * time.Millisecond) // Allow retryJoin to execute

		assert.True(t, raft.initCompleted.Load(), "expected initCompleted to be set to true after successful join")
		mockAPIClient.AssertExpectations(t)
	})

	t.Run("JoinFailsAndRetries", func(t *testing.T) {
		t.Parallel()

		mockAPIClient := new(MockAPIClient)
		mockAPIClient.On("RaftJoin", "node1", "127.0.0.1:8080").Return(errors.New("join error")).Twice()
		mockAPIClient.On("RaftJoin", "node1", "127.0.0.1:8080").Return(nil).Once()
		mockAPIClient.On("Close").Return(nil)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
				Peers:  []string{"http://127.0.0.1:8081"},
			},
			logger: log.Logger,
			getAPIClient: func(_ string) APIClient {
				return mockAPIClient
			},
			initCompleted: atomic.Bool{},
		}

		go raft.retryJoin()

		time.Sleep(3 * time.Second) // Allow retryJoin to execute multiple retries

		assert.True(t, raft.initCompleted.Load(), "expected initCompleted to be set to true after successful join")
		mockAPIClient.AssertExpectations(t)
	})

	t.Run("SkipSelf", func(t *testing.T) {
		t.Parallel()

		mockAPIClient := new(MockAPIClient)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
				Peers:  []string{"http://127.0.0.1:8080"},
			},
			logger: log.Logger,
			getAPIClient: func(_ string) APIClient {
				return mockAPIClient
			},
			initCompleted: atomic.Bool{},
		}

		go raft.retryJoin()

		time.Sleep(100 * time.Millisecond) // Allow retryJoin to execute

		assert.False(t, raft.initCompleted.Load(), "expected initCompleted to remain false when skipping self")
		mockAPIClient.AssertNotCalled(t, "RaftJoin")
	})

	t.Run("InvalidURL", func(t *testing.T) {
		t.Parallel()

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
				Peers:  []string{"http://[::1]:invalid"},
			},
			logger: log.Logger,
			getAPIClient: func(_ string) APIClient {
				return nil // Should not be called
			},
			initCompleted: atomic.Bool{},
		}

		assert.NotPanics(t, func() {
			go raft.retryJoin()
			time.Sleep(100 * time.Millisecond) // Allow retryJoin to execute
		}, "expected retryJoin to not panic for invalid URL")

		assert.False(t, raft.initCompleted.Load(), "expected initCompleted to remain false for invalid URL")
	})
}
func TestForget(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulForget", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockIndexFuture := new(MockIndexFuture)
		mockIndexFuture.On("Error").Return(nil)

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("RemoveServer", hraft.ServerID("node1"), uint64(0), time.Duration(0)).Return(mockIndexFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Forget("node1")
		require.NoError(t, err, "expected Forget to succeed")
		mockRaft.AssertExpectations(t)
		mockIndexFuture.AssertExpectations(t)
	})

	t.Run("NotLeader", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Follower)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Forget("node1")
		require.ErrorIs(t, err, ErrNotALeader, "expected Forget to fail with ErrNotALeader")
		mockRaft.AssertExpectations(t)
	})

	t.Run("RemoveServerError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockIndexFuture := new(MockIndexFuture)
		mockIndexFuture.On("Error").Return(errors.New("remove server error"))

		mockRaft.On("State").Return(hraft.Leader)
		mockRaft.On("RemoveServer", hraft.ServerID("node1"), uint64(0), time.Duration(0)).Return(mockIndexFuture)

		raft := &Raft{
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Forget("node1")
		require.Error(t, err, "expected Forget to fail due to RemoveServer error")
		mockRaft.AssertExpectations(t)
		mockIndexFuture.AssertExpectations(t)
	})
}
func TestJoin(t *testing.T) {
	t.Parallel()

	t.Run("SuccessfulJoin", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockIndexFuture := new(MockIndexFuture)

		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("AddVoter", hraft.ServerID("node2"), hraft.ServerAddress("127.0.0.1:8081"), uint64(0), time.Duration(0)).Return(mockIndexFuture)
		mockIndexFuture.On("Error").Return(nil)
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node2", "127.0.0.1:8081")
		require.NoError(t, err, "expected Join to succeed")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
		mockIndexFuture.AssertExpectations(t)
	})

	t.Run("AlreadyMember", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)

		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:      hraft.ServerID("node2"),
					Address: hraft.ServerAddress("127.0.0.1:8081"),
				},
			},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node2", "127.0.0.1:8081")
		require.NoError(t, err, "expected Join to succeed for already existing member")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
	})

	t.Run("RemoveExistingNode", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockIndexFuture := new(MockIndexFuture)

		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:      hraft.ServerID("node2"),
					Address: hraft.ServerAddress("127.0.0.1:8082"),
				},
			},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("RemoveServer", hraft.ServerID("node2"), uint64(0), time.Duration(0)).Return(mockIndexFuture)
		mockIndexFuture.On("Error").Return(nil)
		mockRaft.On("AddVoter", hraft.ServerID("node2"), hraft.ServerAddress("127.0.0.1:8081"), uint64(0), time.Duration(0)).Return(mockIndexFuture)
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node2", "127.0.0.1:8081")
		require.NoError(t, err, "expected Join to succeed after removing existing node")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
		mockIndexFuture.AssertExpectations(t)
	})

	t.Run("RemoveServerError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockIndexFuture := new(MockIndexFuture)

		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{
				{
					ID:      hraft.ServerID("node2"),
					Address: hraft.ServerAddress("127.0.0.1:8081"),
				},
			},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("RemoveServer", hraft.ServerID("node2"), uint64(0), time.Duration(0)).Return(mockIndexFuture)
		mockIndexFuture.On("Error").Return(errors.New("remove server error"))
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node3", "127.0.0.1:8081")
		require.Error(t, err, "expected Join to fail due to RemoveServer error")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
		mockIndexFuture.AssertExpectations(t)
	})

	t.Run("NotLeader", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockRaft.On("State").Return(hraft.Follower)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node2", "127.0.0.1:8081")
		require.ErrorIs(t, err, ErrNotALeader, "expected Join to fail when not a leader")
		mockRaft.AssertExpectations(t)
	})

	t.Run("ConfigurationError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)

		mockConfigFuture.On("Error").Return(errors.New("configuration error"))
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node2", "127.0.0.1:8081")
		require.Error(t, err, "expected Join to fail due to configuration error")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
	})

	t.Run("AddVoterError", func(t *testing.T) {
		t.Parallel()

		mockRaft := new(MockHRaft)
		mockConfigFuture := new(MockConfigurationFuture)
		mockIndexFuture := new(MockIndexFuture)

		mockConfigFuture.On("Error").Return(nil)
		mockConfigFuture.On("Configuration").Return(hraft.Configuration{
			Servers: []hraft.Server{},
		})
		mockRaft.On("GetConfiguration").Return(mockConfigFuture)
		mockRaft.On("AddVoter", hraft.ServerID("node2"), hraft.ServerAddress("127.0.0.1:8081"), uint64(0), time.Duration(0)).Return(mockIndexFuture)
		mockIndexFuture.On("Error").Return(errors.New("add voter error"))
		mockRaft.On("State").Return(hraft.Leader)

		raft := &Raft{
			config: &Config{
				NodeID: "node1",
				Addr:   "127.0.0.1:8080",
			},
			raftInstance: mockRaft,
			logger:       log.Logger,
		}

		err := raft.Join("node2", "127.0.0.1:8081")
		require.Error(t, err, "expected Join to fail due to AddVoter error")
		mockRaft.AssertExpectations(t)
		mockConfigFuture.AssertExpectations(t)
		mockIndexFuture.AssertExpectations(t)
	})
}
