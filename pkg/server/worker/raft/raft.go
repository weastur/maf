package raft

import (
	"sync"

	"github.com/rs/zerolog/log"
	sentryUtils "github.com/weastur/maf/pkg/utils/sentry"
)

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

func (r *Raft) Run(wg *sync.WaitGroup) {
	log.Info().Msg("Running raft worker")

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentryUtils.Recover(sentryUtils.Fork("fiber"))

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
