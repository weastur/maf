package raft

import (
	"os"
	"sync"

	"github.com/rs/zerolog/log"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

const datadirPerms = 0o700

type Config struct {
	Addr    string
	NodeID  string
	Devmode bool
	Peers   []string
	Datadir string
}

type Raft struct {
	config *Config
	done   chan struct{}
}

func New(config *Config) *Raft {
	log.Trace().Msg("Configuring raft worker")

	return &Raft{config: config, done: make(chan struct{})}
}

func (r *Raft) init() {
	log.Trace().Msg("Initializing raft worker")

	if r.config.Datadir != "" {
		log.Info().Msgf("Using raft data directory: %s", r.config.Datadir)

		if err := os.MkdirAll(r.config.Datadir, datadirPerms); err != nil {
			log.Fatal().Err(err).Msg("Failed to create raft data directory")
		}
	}
}

func (r *Raft) Run(wg *sync.WaitGroup) {
	log.Info().Msg("Running raft worker")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

		r.init()

		<-r.done
	}()
}

func (r *Raft) Stop() {
	log.Info().Msg("Stopping raft worker")

	close(r.done)
}

func (r *Raft) IsLeader() bool {
	return true
}
