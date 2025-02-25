package cmd

import "time"

const (
	defaultHTTPReadTimeout  = 5 * time.Second
	defaultHTTPWriteTimeout = 5 * time.Second
	defaultHTTPIdleTimeout  = 60 * time.Second
)
