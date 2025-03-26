package raft

import (
	"maps"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/internal/utils/logging"
)

type Mapping map[string]string

type SafeStorage struct {
	mu     sync.RWMutex
	data   Mapping
	logger zerolog.Logger
}

func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		mu:     sync.RWMutex{},
		data:   make(Mapping),
		logger: log.With().Str(logging.ComponentCtxKey, "raft-safestorage").Logger(),
	}
}

func (s *SafeStorage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]

	s.logger.Trace().Msgf("Getting from storage: %s:%s", key, val)

	return val, ok
}

func (s *SafeStorage) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Trace().Msgf("Setting in storage: %s:%s", key, value)

	s.data[key] = value
}

func (s *SafeStorage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Trace().Msgf("Deleting from storage: %s", key)

	delete(s.data, key)
}

func (s *SafeStorage) Snapshot() Mapping {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.logger.Trace().Msg("Creating snapshot")

	clone := make(map[string]string)
	maps.Copy(clone, s.data)

	return clone
}

func (s *SafeStorage) Restore(data Mapping) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Trace().Msg("Restoring snapshot")

	maps.Copy(s.data, data)
}
