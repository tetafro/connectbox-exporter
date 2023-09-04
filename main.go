// ConnectBox exporter reads metrics from ConnectBox router using HTTP API,
// and returns them in Prometheus format.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	configFile := flag.String("config", "./config.yaml", "path to config file")
	flag.Parse()

	config, err := ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Create a client for each target
	targets := map[string]*ConnectBox{}
	for _, t := range config.Targets {
		client, err := NewConnectBox(
			t.Addr,
			t.Username,
			t.Password,
		)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to init ConnectBox client: %v", err)
		}
		targets[t.Addr] = client
	}

	// Init prometheus metrics collector
	collector := NewCollector(targets)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/probe", collector)
	//nolint:gosec
	srv := http.Server{
		Addr:    config.ListenAddr,
		Handler: mux,
	}

	// Run HTTP server
	go func() {
		log.Printf("Listening on %s...", config.ListenAddr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for SIGTERM/SIGINT
	<-ctx.Done()

	// Shutdown gracefully
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	fmt.Println("Shutdown gracefully")
}
