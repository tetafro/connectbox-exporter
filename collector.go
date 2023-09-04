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
		http400(w, "Unknown target")
		return
	}

	if err := client.Login(r.Context()); err != nil {
		log.Printf("Failed to login: %v", err)
		http500(w, "Collector error")
		return
	}

	reg := prometheus.NewRegistry()
	c.collectCMState(r.Context(), reg, client)

	if err := client.Logout(r.Context()); err != nil {
		log.Fatalf("Failed to logout: %v", err)
	}

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func (c *Collector) collectCMState(
	ctx context.Context,
	reg *prometheus.Registry,
	client *ConnectBox,
) {
	tunnerTemperatureGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connect_box_tunner_temperature",
		Help: "Tunner temperature.",
	})
	temperatureGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connect_box_temperature",
		Help: "Temperature.",
	})
	operStateGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connect_box_oper_state",
		Help: "Operational state.",
	})
	wanIPv4AddrGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_wan_ipv4_addr",
		Help: "WAN IPv4 address.",
	}, []string{"ip"})
	wanIPv6AddrGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_wan_ipv6_addr",
		Help: "WAN IPv6 address.",
	}, []string{"ip"})

	reg.MustRegister(tunnerTemperatureGauge)
	reg.MustRegister(temperatureGauge)
	reg.MustRegister(operStateGauge)
	reg.MustRegister(wanIPv4AddrGauge)
	reg.MustRegister(wanIPv6AddrGauge)

	var cmstate CMState
	err := client.GetMetrics(ctx, FnCMState, &cmstate)
	if err == nil {
		tunnerTemperatureGauge.Set(float64(cmstate.TunnerTemperature))
		temperatureGauge.Set(float64(cmstate.Temperature))
		var operStateValue float64
		if cmstate.OperState == OperStateOK {
			operStateValue = 1
		}
		operStateGauge.Set(operStateValue)
		wanIPv4AddrGauge.WithLabelValues(cmstate.WANIPv4Addr).Set(1)
		for _, addr := range cmstate.WANIPv6Addrs {
			wanIPv6AddrGauge.WithLabelValues(addr).Set(1)
		}
	} else {
		log.Printf("Failed to get CMState: %v", err)
	}
}

func http400(w http.ResponseWriter, resp string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(resp)) //nolint:errcheck,gosec
}

func http500(w http.ResponseWriter, resp string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(resp)) //nolint:errcheck,gosec
}
