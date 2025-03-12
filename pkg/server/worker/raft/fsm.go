package raft

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"

	"github.com/hashicorp/raft"
)

type Storage interface {
	Get(key string) (string, bool)
	Set(key, value string)
	Delete(key string)
	Snapshot() KeyValue
	Restore(data KeyValue)
}

type FSM struct {
	storage Storage
}

type FSMSnapshot KeyValue

func NewFSM(storage Storage) *FSM {
	return &FSM{storage: storage}
}

func (f *FSM) Apply(rlog *raft.Log) any {
	var cmd Command
	if err := json.Unmarshal(rlog.Data, &cmd); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal command")
	}

	switch cmd.Op {
	case OpSet:
		f.storage.Set(cmd.Key, cmd.Value)
	case OpDelete:
		f.storage.Delete(cmd.Key)
	default:
		log.Fatal().Msgf("unrecognized command %d", cmd.Op)
	}

	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	log.Trace().Msg("Creating snapshot")

	snapshot := FSMSnapshot(f.storage.Snapshot())

	return &snapshot, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	log.Trace().Msg("Restoring snapshot")

	data := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		log.Error().Err(err).Msg("failed to decode snapshot")

		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	f.storage.Restore(KeyValue(data))

	return nil
}

func (fs *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Trace().Msg("Persisting snapshot")

	err := func() error {
		log.Trace().Msg("Encode data")

		data, err := json.Marshal(fs)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal snapshot")

			return fmt.Errorf("failed to marshal snapshot: %w", err)
		}

		if _, err := sink.Write(data); err != nil {
			log.Error().Err(err).Msg("failed to write snapshot")

			return fmt.Errorf("failed to write snapshot: %w", err)
		}

		if err := sink.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close sink")

			return fmt.Errorf("failed to close sink: %w", err)
		}

		return nil
	}()
	if err != nil {
		if err2 := sink.Cancel(); err2 != nil {
			log.Error().Err(err2).Msg("failed to cancel sink")

			return fmt.Errorf("failed to cancel sink: %w", err2)
		}
	}

	return err
}

func (fs *FSMSnapshot) Release() {
	log.Trace().Msg("Releasing snapshot")
}
