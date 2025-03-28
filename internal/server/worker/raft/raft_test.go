package raft

import (
	hraft "github.com/hashicorp/raft"
	"github.com/stretchr/testify/mock"
)

type MockRaftInstance struct {
	mock.Mock
	state hraft.RaftState
}

func (m *MockRaftInstance) State() hraft.RaftState {
	return m.state
}

func (m *MockRaftInstance) Shutdown() *hraft.Future {
	args := m.Called()

	return args.Get(0).(*hraft.Future)
}

func (m *MockRaftInstance) LeadershipTransfer() *hraft.Future {
	args := m.Called()

	return args.Get(0).(*hraft.Future)
}

// func TestNew(t *testing.T) {
// 	config := &Config{
// 		Addr:    "127.0.0.1:8080",
// 		NodeID:  "node1",
// 		Devmode: true,
// 	}

// 	raft := New(config)

// 	assert.NotNil(t, raft)
// 	assert.Equal(t, config, raft.config)
// 	assert.NotNil(t, raft.done)
// 	assert.NotNil(t, raft.logger)
// 	assert.NotNil(t, raft.hlogger)
// 	assert.Empty(t, raft.leadershipChangesChannels)
// }

// func TestIsReady(t *testing.T) {
// 	raft := &Raft{}
// 	assert.False(t, raft.IsReady())

// 	raft.initCompleted.Store(true)
// 	assert.True(t, raft.IsReady())
// }

// func TestIsLive(t *testing.T) {
// 	config := &Config{
// 		Addr:      "127.0.0.1:8080",
// 		NodeID:    "node1",
// 		Devmode:   true,
// 		Datadir:   t.TempDir(),
// 		Bootstrap: true,
// 	}

// 	raft := New(config)
// 	defer raft.Stop()

// 	raft.init()

// 	assert.True(t, raft.initCompleted.Load())
// 	assert.NotNil(t, raft.done)
// 	time.Sleep(3 * time.Second) // Simulate some processing time
// 	assert.True(t, raft.IsLive())
// }

// func TestInit(t *testing.T) {
// 	t.Run("Successful Initialization with bootstrap", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Datadir:   t.TempDir(),
// 			Bootstrap: true,
// 		}

// 		raft := New(config)

// 		raft.init()
// 		defer raft.Stop()

// 		assert.True(t, raft.initCompleted.Load())
// 		assert.NotNil(t, raft.done)

// 		// second bootstrap should not be an issue
// 		time.Sleep(3 * time.Second) // Simulate some processing time
// 		raft.bootstrap()
// 	})

// 	t.Run("Initialization with devmode false", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   false,
// 			Datadir:   t.TempDir(),
// 			Bootstrap: true,
// 		}

// 		raft := New(config)

// 		raft.init()
// 		defer raft.Stop()

// 		assert.True(t, raft.initCompleted.Load())
// 		assert.NotNil(t, raft.done)
// 	})
// }

// func TestRun(t *testing.T) {
// 	config := &Config{
// 		Addr:      "127.0.0.1:8080",
// 		NodeID:    "node1",
// 		Devmode:   true,
// 		Datadir:   t.TempDir(),
// 		Bootstrap: true,
// 	}

// 	raft := New(config)

// 	wg := sync.WaitGroup{}

// 	go func() {
// 		time.Sleep(3 * time.Second) // Simulate some processing time
// 		raft.Stop()
// 	}()

// 	raft.Run(&wg)
// 	wg.Wait()

// 	assert.True(t, raft.initCompleted.Load(), "Raft should be initialized")
// }

// func TestBroadcastLeadershipChange(t *testing.T) {
// 	t.Run("Broadcasts to all channels", func(t *testing.T) {
// 		raft := &Raft{
// 			leadershipChangesChannels: make([]LeadershipChangesCh, 2),
// 		}

// 		ch1 := make(LeadershipChangesCh, 1)
// 		ch2 := make(LeadershipChangesCh, 1)

// 		raft.leadershipChangesChannels[0] = ch1
// 		raft.leadershipChangesChannels[1] = ch2

// 		raft.broadcastLeadershipChange(true)

// 		assert.True(t, <-ch1)
// 		assert.True(t, <-ch2)
// 	})

// 	t.Run("Skips full channels", func(t *testing.T) {
// 		raft := &Raft{
// 			leadershipChangesChannels: make([]LeadershipChangesCh, 2),
// 			logger:                    zerolog.Nop(),
// 		}

// 		ch1 := make(LeadershipChangesCh, 1)
// 		ch2 := make(LeadershipChangesCh, 1)

// 		// Fill ch2 to simulate a full channel
// 		ch2 <- false

// 		raft.leadershipChangesChannels[0] = ch1
// 		raft.leadershipChangesChannels[1] = ch2

// 		raft.broadcastLeadershipChange(true)

// 		assert.True(t, <-ch1)
// 		assert.False(t, <-ch2) // ch2 should remain unchanged
// 	})
// }

// func TestSubscribeOnLeadershipChanges(t *testing.T) {
// 	t.Run("Successfully subscribes to leadership changes", func(t *testing.T) {
// 		raft := &Raft{
// 			leadershipChangesChannels: make([]LeadershipChangesCh, 0),
// 			logger:                    zerolog.Nop(),
// 		}

// 		ch := make(LeadershipChangesCh, 1)
// 		raft.SubscribeOnLeadershipChanges(ch)

// 		assert.Len(t, raft.leadershipChangesChannels, 1)
// 		assert.Equal(t, ch, raft.leadershipChangesChannels[0])
// 	})

// 	t.Run("Multiple subscriptions are handled correctly", func(t *testing.T) {
// 		raft := &Raft{
// 			leadershipChangesChannels: make([]LeadershipChangesCh, 0),
// 			logger:                    zerolog.Nop(),
// 		}

// 		ch1 := make(LeadershipChangesCh, 1)
// 		ch2 := make(LeadershipChangesCh, 1)

// 		raft.SubscribeOnLeadershipChanges(ch1)
// 		raft.SubscribeOnLeadershipChanges(ch2)

// 		assert.Len(t, raft.leadershipChangesChannels, 2)
// 		assert.Equal(t, ch1, raft.leadershipChangesChannels[0])
// 		assert.Equal(t, ch2, raft.leadershipChangesChannels[1])
// 	})
// }

// func TestGet(t *testing.T) {
// 	t.Run("Key exists in storage", func(t *testing.T) {
// 		mockStorage := &MockStorage{data: map[string]string{"existingKey": "value"}}

// 		raft := &Raft{
// 			storage: mockStorage,
// 			logger:  zerolog.Nop(),
// 		}

// 		value, found := raft.Get("existingKey")

// 		assert.True(t, found)
// 		assert.Equal(t, "value", value)
// 	})

// 	t.Run("Key does not exist in storage", func(t *testing.T) {
// 		mockStorage := &MockStorage{data: map[string]string{}}

// 		raft := &Raft{
// 			storage: mockStorage,
// 			logger:  zerolog.Nop(),
// 		}

// 		value, found := raft.Get("missingKey")

// 		assert.False(t, found)
// 		assert.Empty(t, value)
// 	})
// }

// func TestSet(t *testing.T) {
// 	t.Run("Set succeeds when leader", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: true,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft := New(config)
// 		raft.init()

// 		defer raft.Stop()
// 		time.Sleep(3 * time.Second)

// 		err := raft.Set("key", "value")
// 		require.NoError(t, err)

// 		time.Sleep(1 * time.Second)

// 		value, found := raft.Get("key")
// 		assert.True(t, found)
// 		assert.Equal(t, "value", value)
// 	})

// 	t.Run("Set not fails when not leader", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: false,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft := New(config)
// 		raft.init()

// 		defer raft.Stop()
// 		time.Sleep(3 * time.Second)

// 		err := raft.Set("key", "value")
// 		require.NoError(t, err)

// 		time.Sleep(1 * time.Second)

// 		_, found := raft.Get("key")
// 		assert.False(t, found)
// 	})
// }

// func TestDelete(t *testing.T) {
// 	t.Run("Delete succeeds when leader", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: true,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft := New(config)
// 		raft.init()

// 		defer raft.Stop()
// 		time.Sleep(3 * time.Second)

// 		// Set a key first
// 		err := raft.Set("key", "value")
// 		require.NoError(t, err)

// 		// Delete the key
// 		err = raft.Delete("key")
// 		require.NoError(t, err)

// 		// Verify the key is deleted
// 		time.Sleep(1 * time.Second)

// 		_, found := raft.Get("key")
// 		assert.False(t, found)
// 	})

// 	t.Run("Delete does nothing when not leader", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: false,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft := New(config)
// 		raft.init()

// 		defer raft.Stop()
// 		time.Sleep(3 * time.Second)

// 		err := raft.Delete("key")
// 		require.NoError(t, err)
// 	})
// }
// func TestGetInfo(t *testing.T) {
// 	t.Run("Successfully retrieves raft info with verbose=false", func(t *testing.T) {
// 		config := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: true,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft := New(config)
// 		raft.init()

// 		defer raft.Stop()
// 		time.Sleep(3 * time.Second)

// 		info, err := raft.GetInfo(false)
// 		require.NoError(t, err)
// 		require.NotNil(t, info)

// 		assert.Equal(t, "node1", info.ID)
// 		assert.Equal(t, "127.0.0.1:8080", info.Addr)
// 		assert.Equal(t, hraft.Leader.String(), info.State)
// 		assert.Len(t, info.Servers, 1)
// 		assert.Nil(t, info.Stats)

// 		assert.Equal(t, "node1", info.Servers[0].ID)
// 		assert.Equal(t, "127.0.0.1:8080", info.Servers[0].Address)
// 		assert.Equal(t, hraft.Voter.String(), info.Servers[0].Suffrage)
// 		assert.True(t, info.Servers[0].Leader)
// 	})
// }

// func TestJoin(t *testing.T) {
// 	t.Run("Successfully join", func(t *testing.T) {
// 		config1 := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: true,
// 			Datadir:   t.TempDir(),
// 		}
// 		config2 := &Config{
// 			Addr:      "127.0.0.1:8081",
// 			NodeID:    "node2",
// 			Devmode:   true,
// 			Bootstrap: false,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft1 := New(config1)
// 		raft2 := New(config2)
// 		raft1.init()
// 		raft2.init()

// 		defer raft1.Stop()
// 		defer raft2.Stop()
// 		time.Sleep(3 * time.Second)

// 		err := raft1.Join("node2", "127.0.1:8081")
// 		time.Sleep(3 * time.Second)
// 		require.NoError(t, err)

// 		cfg := raft1.raftInstance.GetConfiguration().Configuration()
// 		assert.Len(t, cfg.Servers, 2)
// 	})
// 	t.Run("Join NOT fails when server already exists", func(t *testing.T) {
// 		config1 := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: true,
// 			Datadir:   t.TempDir(),
// 		}
// 		config2 := &Config{
// 			Addr:      "127.0.0.1:8081",
// 			NodeID:    "node2",
// 			Devmode:   true,
// 			Bootstrap: false,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft1 := New(config1)
// 		raft2 := New(config2)
// 		raft1.init()
// 		raft2.init()

// 		defer raft1.Stop()
// 		defer raft2.Stop()
// 		time.Sleep(3 * time.Second)

// 		// First join should succeed
// 		err := raft1.Join("node2", "127.0.0.1:8081")
// 		require.NoError(t, err)

// 		// Attempt to join the same server again
// 		err = raft1.Join("node2", "127.0.0.1:8081")
// 		require.NoError(t, err, "Joining an existing server should not return an error but should be ignored")

// 		cfg := raft1.raftInstance.GetConfiguration().Configuration()
// 		assert.Len(t, cfg.Servers, 2)
// 	})

// 	t.Run("Re-join with a new ID", func(t *testing.T) {
// 		config1 := &Config{
// 			Addr:      "127.0.0.1:8080",
// 			NodeID:    "node1",
// 			Devmode:   true,
// 			Bootstrap: true,
// 			Datadir:   t.TempDir(),
// 		}
// 		config2 := &Config{
// 			Addr:      "127.0.0.1:8081",
// 			NodeID:    "node2",
// 			Devmode:   true,
// 			Bootstrap: false,
// 			Datadir:   t.TempDir(),
// 		}

// 		raft1 := New(config1)
// 		raft2 := New(config2)
// 		raft1.init()
// 		raft2.init()

// 		defer raft1.Stop()
// 		defer raft2.Stop()
// 		time.Sleep(3 * time.Second)

// 		// First join with the original ID
// 		err := raft1.Join("node2", "127.0.0.1:8081")
// 		require.NoError(t, err)

// 		// Re-join with a new ID
// 		err = raft1.Join("node3", "127.0.0.1:8081")
// 		require.NoError(t, err, "Re-joining with a new ID should succeed")

// 		cfg := raft1.raftInstance.GetConfiguration().Configuration()
// 		assert.Len(t, cfg.Servers, 2)
// 	})
// }

// // func TestForget(t *testing.T) {
// // 	t.Run("Successfully forget server", func(t *testing.T) {
// // 		config1 := &Config{
// // 			Addr:      "127.0.0.1:8080",
// // 			NodeID:    "node1",
// // 			Devmode:   true,
// // 			Bootstrap: true,
// // 			Datadir:   t.TempDir(),
// // 		}
// // 		config2 := &Config{
// // 			Addr:      "127.0.0.1:8081",
// // 			NodeID:    "node2",
// // 			Devmode:   true,
// // 			Bootstrap: false,
// // 			Datadir:   t.TempDir(),
// // 		}

// // 		raft1 := New(config1)
// // 		raft2 := New(config2)
// // 		raft1.init()
// // 		raft2.init()
// // 		time.Sleep(3 * time.Second)

// // 		defer raft1.Stop()
// // 		defer raft2.Stop()

// // 		err := raft1.Join("node2", "127.0.1:8081")
// // 		time.Sleep(3 * time.Second)
// // 		require.NoError(t, err)

// // 		cfg := raft1.raftInstance.GetConfiguration().Configuration()
// // 		assert.Len(t, cfg.Servers, 2)

// // 		err = raft1.Forget("node1")
// // 		require.NoError(t, err)
// // 		time.Sleep(3 * time.Second)
// // 		cfg = raft1.raftInstance.GetConfiguration().Configuration()
// // 		assert.Len(t, cfg.Servers, 1)
// // 	})
// // }
