package raft

import (
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/weastur/maf/pkg/utils/logging"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

const (
	datadirPerms     = 0o700
	transportMaxPool = 3
	transportTimeout = 10 * time.Second
	snapshotRetain   = 2
)

type Config struct {
	Addr    string
	NodeID  string
	Devmode bool
	Peers   []string
	Datadir string
}

type Raft struct {
	config      *Config
	hrconfig    *raft.Config
	done        chan struct{}
	logger      zerolog.Logger
	hrlogger    *HCZeroLogger
	hrtransport *raft.NetworkTransport
	hrsnapshots *raft.FileSnapshotStore
}

func New(config *Config) *Raft {
	log.Trace().Msg("Configuring raft worker")

	return &Raft{
		config:   config,
		done:     make(chan struct{}),
		logger:   log.With().Str(logging.ComponentCtxKey, "raft").Logger(),
		hrlogger: NewHCZeroLogger(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
	}
}

func (r *Raft) init() {
	r.logger.Trace().Msg("Initializing")

	r.ensureDatadir()
	r.configureHRaft()
	r.configureHRTransport()
	r.configureHRSnapshots()

	var logStore raft.LogStore

	var stableStore raft.StableStore

	if r.config.Devmode {
		logStore = raft.NewInmemStore()
		stableStore = raft.NewInmemStore()
	} else {
		boltDB, err := raftboltdb.New(raftboltdb.Options{
			Path: filepath.Join(r.config.Datadir, "raft.db"),
		})
		if err != nil {
			r.logger.Fatal().Err(err).Msg("Failed to create boltDB")
		}

		logStore = boltDB
		stableStore = boltDB
	}

	fsm := NewFSM(NewSafeStorage())

	ra, err := raft.NewRaft(r.hrconfig, fsm, logStore, stableStore, r.hrsnapshots, r.hrtransport)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to create raft")
	}

	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      r.hrconfig.LocalID,
				Address: r.hrtransport.LocalAddr(),
			},
		},
	}
	ra.BootstrapCluster(configuration)
}

func (r *Raft) configureHRSnapshots() {
	var err error

	r.hrsnapshots, err = raft.NewFileSnapshotStore(r.config.Datadir, snapshotRetain, os.Stderr)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to create snapshot store")
	}
}

func (r *Raft) configureHRTransport() {
	addr, err := net.ResolveTCPAddr("tcp", r.config.Addr)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to resolve TCP address")
	}

	r.hrtransport, err = raft.NewTCPTransport(r.config.Addr, addr, transportMaxPool, transportTimeout, os.Stderr)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to create TCP transport")
	}
}

func (r *Raft) configureHRaft() {
	r.logger.Trace().Msg("Configuring raft")

	r.hrconfig = raft.DefaultConfig()
	r.hrconfig.LocalID = raft.ServerID(r.config.NodeID)
	r.hrconfig.Logger = r.hrlogger
}

func (r *Raft) ensureDatadir() {
	if r.config.Datadir != "" {
		r.logger.Info().Msgf("Using raft data directory: %s", r.config.Datadir)

		if err := os.MkdirAll(r.config.Datadir, datadirPerms); err != nil {
			r.logger.Fatal().Err(err).Msg("Failed to create raft data directory")
		}
	}
}

func (r *Raft) Run(wg *sync.WaitGroup) {
	r.logger.Info().Msg("Running")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

		r.init()

		<-r.done
	}()
}

func (r *Raft) Stop() {
	r.logger.Info().Msg("Stopping")

	close(r.done)
}

func (r *Raft) IsLeader() bool {
	return true
}
