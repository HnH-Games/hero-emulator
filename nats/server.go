package nats

import (
	"time"

	"github.com/nats-io/gnatsd/server"
)

// DefaultTestOptions are default options for the unit tests.
var DefaultOptions = server.Options{
	Host:           "127.0.0.1",
	Port:           4222,
	NoLog:          false,
	NoSigs:         false,
	MaxControlLine: 256,
}

// RunServer starts a new Go routine based server
func RunServer(opts *server.Options) *server.Server {
	if opts == nil {
		opts = &DefaultOptions
	}

	s := server.New(opts)
	if s == nil {
		panic("No NATS Server object returned.")
	}

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		panic("Unable to start NATS Server in Go Routine")
	}
	return s
}
