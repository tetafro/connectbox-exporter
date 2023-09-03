package main

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector collects metrics from a remote ConnectBox router.
type Collector struct {
	targets map[string]*ConnectBox
}

// NewCollector creates new collector.
func NewCollector(targets map[string]*ConnectBox) *Collector {
	return &Collector{targets: targets}
}

// ServeHTTP handles requests from Prometheus. It collects all metrics,
// writes them to a temporary registry, and then returns.
func (c *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	client, ok := c.targets[target]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unknown target")) //nolint:errcheck,gosec
		return
	}

	reg := prometheus.NewRegistry()
	c.collect(r.Context(), reg, client)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func (c *Collector) collect(
	ctx context.Context,
	reg *prometheus.Registry,
	client *ConnectBox,
) {
	temperatureGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connect_box_temperature",
		Help: "Temperature.",
	})
	reg.MustRegister(temperatureGauge)

	if err := client.Login(ctx); err != nil {
		log.Fatalf("ERROR: Failed to login: %v", err)
	}
	log.Print("Logged in successfully")

	var cmstate CMState
	err := client.GetMetrics(ctx, FnCMState, &cmstate)
	if err == nil {
		temperatureGauge.Set(float64(cmstate.Temperature))
	} else {
		log.Printf("ERROR: Failed to get CMState: %v", err)
	}

	if err := client.Logout(ctx); err != nil {
		log.Fatalf("ERROR: Failed to logout: %v", err)
	}
	log.Print("Logged out successfully")
}
