package raft

import (
	"maps"
	"sync"

	"github.com/rs/zerolog/log"
)

type KeyValue map[string]string

type SafeStorage struct {
	mu   sync.RWMutex
	data KeyValue
}

func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		mu:   sync.RWMutex{},
		data: make(KeyValue),
	}
}

func (s *SafeStorage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]

	log.Trace().Msgf("Getting from storage: %s:%s", key, val)

	return val, ok
}

func (s *SafeStorage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Trace().Msgf("Setting in storage: %s:%s", key, value)

	s.data[key] = value
}

func (s *SafeStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Trace().Msgf("Deleting from storage: %s", key)

	delete(s.data, key)
}

func (s *SafeStorage) Snapshot() KeyValue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	log.Trace().Msg("Creating snapshot")

	clone := make(map[string]string)
	maps.Copy(clone, s.data)

	return clone
}

func (s *SafeStorage) Restore(data KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Trace().Msg("Restoring snapshot")

	maps.Copy(s.data, data)
}
