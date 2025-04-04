package raft

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	neturl "net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-hclog"
	hraft "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	hclogzerolog "github.com/weastur/hclog-zerolog"
	apiClient "github.com/weastur/maf/internal/server/client"
	"github.com/weastur/maf/internal/utils/logging"
)

const (
	datadirPerms     = 0o700
	transportMaxPool = 3
	transportTimeout = 10 * time.Second
	retryJoinDelay   = time.Second
	snapshotRetain   = 2
	dbName           = "raft.db"
	cmdTimeout       = 10 * time.Second
)

type LeadershipChangesCh chan bool

type Sentry interface {
	Recover()
}

type Config struct {
	Addr               string
	NodeID             string
	Devmode            bool
	Peers              []string
	Datadir            string
	Bootstrap          bool
	ServerAPITLSConfig *apiClient.TLSConfig
}

type HRaft interface {
	State() hraft.RaftState
	BootstrapCluster(configuration hraft.Configuration) hraft.Future
	LeadershipTransfer() hraft.Future
	Shutdown() hraft.Future
	GetConfiguration() hraft.ConfigurationFuture
	RemoveServer(id hraft.ServerID, prevIndex uint64, timeout time.Duration) hraft.IndexFuture
	AddVoter(id hraft.ServerID, address hraft.ServerAddress, prevIndex uint64, timeout time.Duration) hraft.IndexFuture
	LeaderWithID() (hraft.ServerAddress, hraft.ServerID)
	Stats() map[string]string
	Apply(cmd []byte, timeout time.Duration) hraft.ApplyFuture
	LeaderCh() <-chan bool
}

type APIClient interface {
	RaftJoin(nodeID, addr string) error
	Close() error
}

type Raft struct {
	config                    *Config
	hrconfig                  *hraft.Config
	done                      chan struct{}
	logger                    zerolog.Logger
	hlogger                   hclog.Logger
	transport                 hraft.Transport
	snapshotStore             hraft.SnapshotStore
	logStore                  hraft.LogStore
	stableStore               hraft.StableStore
	fsm                       hraft.FSM
	storage                   Storage
	raftInstance              HRaft
	initCompleted             atomic.Bool
	leadershipChangesChannels []LeadershipChangesCh
	sentry                    Sentry
	getAPIClient              func(string) APIClient
}

func New(config *Config, sentry Sentry) *Raft {
	log.Trace().Msg("Configuring raft worker")

	return &Raft{
		config:                    config,
		done:                      make(chan struct{}),
		logger:                    log.With().Str(logging.ComponentCtxKey, "raft").Logger(),
		hlogger:                   hclogzerolog.New(log.With().Str(logging.ComponentCtxKey, "hraft").Logger()),
		leadershipChangesChannels: make([]LeadershipChangesCh, 0),
		sentry:                    sentry,
		getAPIClient: func(peer string) APIClient {
			return apiClient.NewWithAutoTLS(peer, config.ServerAPITLSConfig, true)
		},
	}
}

func (r *Raft) IsReady() bool {
	return r.initCompleted.Load()
}

func (r *Raft) IsLive() bool {
	return r.raftInstance.State() == hraft.Leader || r.raftInstance.State() == hraft.Follower
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
	r.monitorLeadership()

	if r.config.Bootstrap {
		r.bootstrap()
	} else {
		go r.retryJoin()
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

	switch {
	case errors.Is(err, hraft.ErrCantBootstrap):
		r.logger.Warn().Msg("Can't bootstrap cluster as it already exists. Ignoring")
		r.initCompleted.Store(true)
	case err != nil:
		panic("Failed to bootstrap cluster")
	default:
		r.initCompleted.Store(true)
	}
}

func (r *Raft) retryJoin() {
	r.logger.Info().Msg("Retrying to join peers")

	for {
		for _, peer := range r.config.Peers {
			r.logger.Debug().Msgf("Joining peer %s", peer)

			peerURL, err := neturl.Parse(peer)
			if err != nil {
				r.logger.Warn().Err(err).Msgf("Failed to parse peer URL %s", peer)

				continue
			}

			if peerURL.Host == r.config.Addr {
				r.logger.Debug().Msg("Skipping self")

				continue
			}

			api := r.getAPIClient(peer)
			if err := api.RaftJoin(r.config.NodeID, r.config.Addr); err != nil {
				r.logger.Warn().Err(err).Msgf("Failed to join peer %s", peer)
				api.Close()
			} else {
				r.logger.Info().Msgf("Successfully joined peer %s", peer)
				api.Close()
				r.initCompleted.Store(true)

				return
			}
		}

		time.Sleep(retryJoinDelay)
	}
}

func (r *Raft) initRaftInstance() {
	var err error

	r.raftInstance, err = hraft.NewRaft(r.hrconfig, r.fsm, r.logStore, r.stableStore, r.snapshotStore, r.transport)
	if err != nil {
		panic("Failed to create raft")
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
			panic("Failed to create boltDB")
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
		panic("Failed to create snapshot store")
	}
}

func (r *Raft) initTransport() {
	r.logger.Trace().Msgf("Initializing transport for %s", r.config.Addr)

	addr, err := net.ResolveTCPAddr("tcp", r.config.Addr)
	if err != nil {
		panic("Failed to resolve TCP address")
	}

	r.transport, err = hraft.NewTCPTransportWithLogger(r.config.Addr, addr, transportMaxPool, transportTimeout, r.hlogger)
	if err != nil {
		panic("Failed to create TCP transport")
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
			panic("Failed to create raft data directory")
		}
	}
}

func (r *Raft) Run(wg *sync.WaitGroup) {
	r.logger.Info().Msg("Running")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer r.sentry.Recover()

		r.init()

		<-r.done
	}()
}

func (r *Raft) Stop() {
	r.logger.Info().Msg("Stopping")

	if !r.initCompleted.Load() {
		r.logger.Warn().Msg("Raft not initialized, skipping stop")

		return
	}

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

	if !r.IsLeader() {
		r.logger.Warn().Msg("I'm not a leader, can't proceed with join")

		return ErrNotALeader
	}

	cfgFuture := r.raftInstance.GetConfiguration()
	if err := cfgFuture.Error(); err != nil {
		r.logger.Error().Err(err).Msg("Failed to get raft configuration")

		return fmt.Errorf("failed to get raft configuration: %w", err)
	}

	cfg := cfgFuture.Configuration()

	rNodeID := hraft.ServerID(serverID)
	rAddr := hraft.ServerAddress(addr)

	// Check if the server is already a member of the cluster or needs to be removed first
	// due to address or ID change.
	for _, srv := range cfg.Servers {
		if srv.ID == rNodeID && srv.Address == rAddr {
			r.logger.Info().Msgf("node %s at %s already member of cluster, ignoring join request", serverID, addr)

			return nil
		} else if srv.ID == rNodeID || srv.Address == rAddr {
			r.logger.Info().Msgf("node %s at %s already member of cluster, removing existing node", serverID, addr)

			idxFuture := r.raftInstance.RemoveServer(srv.ID, 0, 0)
			if err := idxFuture.Error(); err != nil {
				r.logger.Err(err).Msgf("Failed to remove existing node %s at %s", srv.ID, srv.Address)

				return fmt.Errorf("failed to remove existing node %s at %s: %w", srv.ID, srv.Address, err)
			}
		}
	}

	idxFuture := r.raftInstance.AddVoter(rNodeID, rAddr, 0, 0)
	if err := idxFuture.Error(); err != nil {
		r.logger.Err(err).Msg("Failed to add voter")

		return fmt.Errorf("failed to add voter: %w", err)
	}

	r.logger.Info().Msgf("Successfully added %s at %s", serverID, addr)

	return nil
}

func (r *Raft) Forget(serverID string) error {
	r.logger.Trace().Msgf("Forget %s", serverID)

	if !r.IsLeader() {
		r.logger.Warn().Msg("I'm not a leader, can't proceed with join")

		return ErrNotALeader
	}

	idxFuture := r.raftInstance.RemoveServer(hraft.ServerID(serverID), 0, 0)
	if err := idxFuture.Error(); err != nil {
		r.logger.Err(err).Msgf("Failed to remove existing node %s", serverID)

		return fmt.Errorf("failed to remove existing node %s: %w", serverID, err)
	}

	return nil
}

func (r *Raft) GetInfo(verbose bool) (*Info, error) {
	r.logger.Trace().Msg("Getting status")

	cfgFuture := r.raftInstance.GetConfiguration()
	if err := cfgFuture.Error(); err != nil {
		r.logger.Error().Err(err).Msg("Failed to get raft configuration")

		return nil, fmt.Errorf("failed to get raft configuration: %w", err)
	}

	cfg := cfgFuture.Configuration()

	info := &Info{
		ID:      r.config.NodeID,
		Addr:    r.config.Addr,
		State:   r.raftInstance.State().String(),
		Servers: make([]Server, 0),
		Stats:   nil,
	}

	lAddr, lID := r.raftInstance.LeaderWithID()

	for _, srv := range cfg.Servers {
		info.Servers = append(info.Servers, Server{
			ID:       string(srv.ID),
			Address:  string(srv.Address),
			Suffrage: srv.Suffrage.String(),
			Leader:   srv.ID == lID && srv.Address == lAddr,
		})
	}

	if verbose {
		info.Stats = r.raftInstance.Stats()
	}

	return info, nil
}

func (r *Raft) Get(key string) (string, bool) {
	r.logger.Trace().Msgf("Getting key %s", key)

	return r.storage.Get(key)
}

func (r *Raft) applyCommand(op OpType, key, value string) error {
	cmd := makeCommand(op, key, value)

	data, err := json.Marshal(cmd)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to marshal command")

		return fmt.Errorf("failed to marshal command: %w", err)
	}

	applyFuture := r.raftInstance.Apply(data, cmdTimeout)

	if err := applyFuture.Error(); err != nil {
		log.Error().Err(err).Msg("failed to apply command")

		return fmt.Errorf("failed to apply command: %w", err)
	}

	return nil
}

func (r *Raft) Set(key, value string) error {
	if !r.IsLeader() {
		r.logger.Warn().Msg("I'm not a leader, can't proceed with set")

		return nil
	}

	return r.applyCommand(OpSet, key, value)
}

func (r *Raft) Delete(key string) error {
	if !r.IsLeader() {
		r.logger.Warn().Msg("I'm not a leader, can't proceed with delete")

		return nil
	}

	return r.applyCommand(OpDelete, key, "")
}

func (r *Raft) SubscribeOnLeadershipChanges(ch LeadershipChangesCh) {
	r.logger.Trace().Msg("Registering leadership changes channel")

	r.leadershipChangesChannels = append(r.leadershipChangesChannels, ch)
}

func (r *Raft) broadcastLeadershipChange(isLeader bool) {
	for _, ch := range r.leadershipChangesChannels {
		select {
		case ch <- isLeader:
		default:
			r.logger.Warn().Msg("Leadership change channel is full, skipping")
		}
	}
}

func (r *Raft) monitorLeadership() {
	r.logger.Info().Msg("Monitoring leadership changes")

	go func() {
		for {
			select {
			case isLeader := <-r.raftInstance.LeaderCh():
				if isLeader {
					r.logger.Info().Msg("Became leader, performing leader-specific tasks")

					r.broadcastLeadershipChange(true)
				} else {
					r.logger.Info().Msg("Lost leadership")

					r.broadcastLeadershipChange(false)
				}
			case <-r.done:
				r.logger.Info().Msg("Stopping leadership monitoring")

				return
			}
		}
	}()
}
