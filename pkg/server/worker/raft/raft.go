package raft

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	hraft "github.com/hashicorp/raft"
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
	dbName           = "raft.db"
)

type Config struct {
	Addr      string
	NodeID    string
	Devmode   bool
	Peers     []string
	Datadir   string
	Bootstrap bool
}

type Raft struct {
	config        *Config
	hrconfig      *hraft.Config
	done          chan struct{}
	logger        zerolog.Logger
	hlogger       hclog.Logger
	transport     hraft.Transport
	snapshotStore hraft.SnapshotStore
	logStore      hraft.LogStore
	stableStore   hraft.StableStore
	fsm           hraft.FSM
	storage       Storage
	raftInstance  *hraft.Raft
}

func New(config *Config) *Raft {
	log.Trace().Msg("Configuring raft worker")

	return &Raft{
		config:  config,
		done:    make(chan struct{}),
		logger:  log.With().Str(logging.ComponentCtxKey, "raft").Logger(),
		hlogger: NewHCZeroLogger(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
	}
}

func (r *Raft) init() {
	r.logger.Trace().Msg("Initializing")

	r.ensureDatadir()
	r.configureRaft()
	r.initTransport()
	r.initSnapshotStore()
	r.initStore()
	r.initFSM()
	r.initRaftInstance()

	if r.config.Bootstrap {
		r.bootstrap()
	}
}

func (r *Raft) bootstrap() {
	configuration := hraft.Configuration{
		Servers: []hraft.Server{
			{
				ID:      r.hrconfig.LocalID,
				Address: r.transport.LocalAddr(),
			},
		},
	}

	r.logger.Info().Msg("Bootstrapping raft cluster with configuration")

	err := r.raftInstance.BootstrapCluster(configuration).Error()
	if errors.Is(err, hraft.ErrCantBootstrap) {
		r.logger.Warn().Msg("Can't bootstrap cluster as it already exists. Ignoring")
	} else if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to bootstrap cluster")
	}
}

func (r *Raft) initRaftInstance() {
	var err error

	r.raftInstance, err = hraft.NewRaft(r.hrconfig, r.fsm, r.logStore, r.stableStore, r.snapshotStore, r.transport)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to create raft")
	}
}

func (r *Raft) initFSM() {
	r.storage = NewSafeStorage()
	r.fsm = NewFSM(r.storage)
}

func (r *Raft) initStore() {
	r.logger.Trace().Msg("Initializing store")

	if r.config.Devmode {
		r.logger.Info().Msg("Using in-memory store")
		r.logStore = hraft.NewInmemStore()
		r.stableStore = hraft.NewInmemStore()
	} else {
		r.logger.Info().Msg("Using boltdb store")

		boltDB, err := raftboltdb.New(raftboltdb.Options{
			Path: filepath.Join(r.config.Datadir, dbName),
		})
		if err != nil {
			r.logger.Fatal().Err(err).Msg("Failed to create boltDB")
		}

		r.logStore = boltDB
		r.stableStore = boltDB
	}
}

func (r *Raft) initSnapshotStore() {
	var err error

	r.logger.Trace().Msg("Initializing snapshot store")

	r.snapshotStore, err = hraft.NewFileSnapshotStoreWithLogger(r.config.Datadir, snapshotRetain, r.hlogger)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to create snapshot store")
	}
}

func (r *Raft) initTransport() {
	r.logger.Trace().Msgf("Initializing transport for %s", r.config.Addr)

	addr, err := net.ResolveTCPAddr("tcp", r.config.Addr)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to resolve TCP address")
	}

	r.transport, err = hraft.NewTCPTransportWithLogger(r.config.Addr, addr, transportMaxPool, transportTimeout, r.hlogger)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("Failed to create TCP transport")
	}
}

func (r *Raft) configureRaft() {
	r.logger.Trace().Msg("Configuring raft")

	r.hrconfig = hraft.DefaultConfig()
	r.logger.Trace().Msgf("Raft server ID: %s", r.config.NodeID)
	r.hrconfig.LocalID = hraft.ServerID(r.config.NodeID)
	r.hrconfig.Logger = r.hlogger
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

	if r.IsLeader() {
		r.logger.Info().Msg("I'm the leader, stepping down")

		if err := r.raftInstance.LeadershipTransfer().Error(); err != nil {
			r.logger.Error().Err(err).Msg("Failed to transfer leadership")
		}
	}

	if err := r.raftInstance.Shutdown().Error(); err != nil {
		r.logger.Error().Err(err).Msg("Failed to shutdown raft")
	}

	close(r.done)
}

func (r *Raft) IsLeader() bool {
	return r.raftInstance.State() == hraft.Leader
}

func (r *Raft) Join(serverID, addr string) error {
	r.logger.Trace().Msgf("Joining %s at %s", serverID, addr)

	return nil
}

func (r *Raft) Leave(serverID string) error {
	r.logger.Trace().Msgf("Leaving %s", serverID)

	return nil
}
