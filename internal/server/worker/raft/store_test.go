package raft

import (
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	storage := NewSafeStorage()

	storage.Set(key, value)

	got, ok := storage.Get(key)
	if !ok {
		t.Fatalf("expected key %s to exist", key)
	}

	if got != value {
		t.Errorf("expected value %s, got %s", value, got)
	}
}

func TestSafeStorage_Delete(t *testing.T) {
	storage := NewSafeStorage()

	storage.Set(key, value)

	storage.Delete(key)

	_, ok := storage.Get(key)
	if ok {
		t.Errorf("expected key %s to be deleted", key)
	}
}

func TestSafeStorage_Snapshot(t *testing.T) {
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
	if !cmp.Equal(snapshot, data) {
		t.Errorf("expected snapshot %v, got %v", data, snapshot)
	}
}

func TestSafeStorage_Restore(t *testing.T) {
	var data Mapping

	storage := NewSafeStorage()

	data = map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	storage.Restore(data)

	for k, v := range data {
		got, ok := storage.Get(k)
		if !ok {
			t.Fatalf("expected key %s to exist", k)
		}

		if got != v {
			t.Errorf("expected value %s, got %s", v, got)
		}
	}
}

func TestSafeStorage_ConcurrentAccess(_ *testing.T) {
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
