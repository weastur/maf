package raft

import (
	"os"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	key   = "key1"
	value = "value1"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	log.Logger = log.Output(zerolog.Nop())

	os.Exit(m.Run())
}

func TestSafeStorage_GetSet(t *testing.T) {
	t.Parallel()

	storage := NewSafeStorage()

	storage.Set(key, value)

	got, ok := storage.Get(key)
	require.True(t, ok, "expected key %s to exist", key)
	assert.Equal(t, value, got, "expected value %s, got %s", value, got)
}

func TestSafeStorage_Delete(t *testing.T) {
	t.Parallel()

	storage := NewSafeStorage()

	storage.Set(key, value)
	storage.Delete(key)

	_, ok := storage.Get(key)
	assert.False(t, ok, "expected key %s to not exist", key)
}

func TestSafeStorage_Snapshot(t *testing.T) {
	t.Parallel()

	var data Mapping

	storage := NewSafeStorage()

	data = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	for k, v := range data {
		storage.Set(k, v)
	}

	snapshot := storage.Snapshot()
	assert.Equal(t, data, snapshot, "expected snapshot %v, got %v", data, snapshot)
}

func TestSafeStorage_Restore(t *testing.T) {
	t.Parallel()

	var data Mapping

	storage := NewSafeStorage()

	data = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	storage.Restore(data)

	for k, v := range data {
		got, ok := storage.Get(k)
		require.True(t, ok, "expected key %s to exist", k)
		assert.Equal(t, v, got, "expected value %s, got %s", v, got)
	}
}

func TestSafeStorage_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	storage := NewSafeStorage()

	writer := func() {
		for range 1000 {
			storage.Set(key, value)
		}
	}

	reader := func() {
		for range 1000 {
			_, _ = storage.Get(key)
		}
	}

	deleter := func() {
		for range 1000 {
			storage.Delete(key)
		}
	}

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		writer()
	}()
	go func() {
		defer wg.Done()
		reader()
	}()
	go func() {
		defer wg.Done()
		deleter()
	}()
	wg.Wait()

	_, _ = storage.Get(key)
}
