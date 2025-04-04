package raft

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/internal/utils/logging"
)

type Storage interface {
	Get(key string) (string, bool)
	Set(key, value string)
	Delete(key string)
	Snapshot() Mapping
	Restore(data Mapping)
}

type FSM struct {
	storage Storage
	logger  zerolog.Logger
}

type FSMSnapshot struct {
	data   Mapping
	logger zerolog.Logger
}

func NewFSM(storage Storage) *FSM {
	return &FSM{
		storage: storage,
		logger:  log.With().Str(logging.ComponentCtxKey, "raft-fsm").Logger(),
	}
}

func (f *FSM) Apply(rlog *raft.Log) any {
	var cmd Command
	if err := json.Unmarshal(rlog.Data, &cmd); err != nil {
		panic("failed to unmarshal command")
	}

	switch cmd.Op {
	case OpSet:
		f.storage.Set(cmd.Key, cmd.Value)
	case OpDelete:
		f.storage.Delete(cmd.Key)
	default:
		panic("unrecognized command " + cmd.Op.String())
	}

	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.logger.Trace().Msg("Creating snapshot")

	return &FSMSnapshot{
		data:   f.storage.Snapshot(),
		logger: f.logger,
	}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	f.logger.Trace().Msg("Restoring snapshot")

	data := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		f.logger.Error().Err(err).Msg("failed to decode snapshot")

		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	f.storage.Restore(Mapping(data))

	return nil
}

func (fs *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	fs.logger.Trace().Msg("Persisting snapshot")

	err := func() error {
		fs.logger.Trace().Msg("Encode data")

		data, err := json.Marshal(fs.data)
		if err != nil {
			fs.logger.Error().Err(err).Msg("failed to marshal snapshot")

			return fmt.Errorf("failed to marshal snapshot: %w", err)
		}

		if _, err := sink.Write(data); err != nil {
			fs.logger.Error().Err(err).Msg("failed to write snapshot")

			return fmt.Errorf("failed to write snapshot: %w", err)
		}

		if err := sink.Close(); err != nil {
			fs.logger.Error().Err(err).Msg("failed to close sink")

			return fmt.Errorf("failed to close sink: %w", err)
		}

		return nil
	}()
	if err != nil {
		if err2 := sink.Cancel(); err2 != nil {
			fs.logger.Error().Err(err2).Msg("failed to cancel sink")

			return fmt.Errorf("failed to cancel sink: %w", err2)
		}
	}

	return err
}

func (fs *FSMSnapshot) Release() {
	fs.logger.Trace().Msg("Releasing snapshot")
}
