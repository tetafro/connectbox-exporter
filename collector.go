package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector collects metrics from a remote ConnectBox router.
type Collector struct {
	targets map[string]MetricsClient
}

// NewCollector creates new collector.
func NewCollector(targets map[string]MetricsClient) *Collector {
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
	defer func() {
		// Use a separate context to avoid cancelling logout when
		// the request is cancelled
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := client.Logout(ctx); err != nil {
			log.Fatalf("Failed to logout: %v", err)
		}
	}()

	// NOTE: Parallel requests are not possible due to how the auth system
	// works - a new token is required for every request
	reg := prometheus.NewRegistry()
	c.collectCMState(r.Context(), reg, client)
	c.collectCMSSystemInfo(r.Context(), reg, client)
	c.collectLANUserTable(r.Context(), reg, client)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func (c *Collector) collectCMSSystemInfo(
	ctx context.Context,
	reg *prometheus.Registry,
	client MetricsClient,
) {
	cmDocsisModeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_cm_docsis_mode",
		Help: "DocSis mode.",
	}, []string{"mode"})
	cmHardwareVersionGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_cm_hardware_version",
		Help: "Hardware_version.",
	}, []string{"version"})
	cmMacAddrGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_cm_mac_addr",
		Help: "MAC address.",
	}, []string{"addr"})
	cmSerialNumberGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_cm_serial_number",
		Help: "Serial number.",
	}, []string{"sn"})
	cmSystemUptimeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_cm_system_uptime",
		Help: "System uptime.",
	}, []string{})
	cmNetworkAccessGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_cm_network_access",
		Help: "Network access.",
	}, []string{})

	reg.MustRegister(cmDocsisModeGauge)
	reg.MustRegister(cmHardwareVersionGauge)
	reg.MustRegister(cmMacAddrGauge)
	reg.MustRegister(cmSerialNumberGauge)
	reg.MustRegister(cmSystemUptimeGauge)
	reg.MustRegister(cmNetworkAccessGauge)

	var data CMSystemInfo
	err := client.GetMetrics(ctx, FnCMSystemInfo, &data)
	if err != nil {
		log.Printf("Failed to get CMSSystemInfo: %v", err)
		return
	}

	cmDocsisModeGauge.WithLabelValues(data.DocsisMode).Set(1)
	cmHardwareVersionGauge.WithLabelValues(data.HardwareVersion).Set(1)
	cmMacAddrGauge.WithLabelValues(data.MacAddr).Set(1)
	cmSerialNumberGauge.WithLabelValues(data.SerialNumber).Set(1)
	cmSystemUptimeGauge.WithLabelValues().Set(float64(data.SystemUptime))
	var val float64
	if data.NetworkAccess == NetworkAccessAllowed {
		val = 1
	}
	cmNetworkAccessGauge.WithLabelValues().Set(val)
}

func (c *Collector) collectLANUserTable(
	ctx context.Context,
	reg *prometheus.Registry,
	client MetricsClient,
) {
	clientGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_lan_client",
		Help: "cm_docsis_mode.",
	}, []string{
		"connection",
		"interface",
		"ipv4_addr",
		"hostname",
		"MACAddr",
	})

	reg.MustRegister(clientGauge)

	var data LANUserTable
	err := client.GetMetrics(ctx, FnLANUserTable, &data)
	if err != nil {
		log.Printf("Failed to get LANUserTable: %v", err)
		return
	}

	for _, c := range data.Ethernet {
		clientGauge.WithLabelValues(
			"ethernet",
			c.Interface,
			c.IPv4Addr,
			c.Hostname,
			c.MACAddr,
		).Set(1)
	}
	for _, c := range data.WIFI {
		clientGauge.WithLabelValues(
			"wifi",
			c.Interface,
			c.IPv4Addr,
			c.Hostname,
			c.MACAddr,
		).Set(1)
	}
}

func (c *Collector) collectCMState(
	ctx context.Context,
	reg *prometheus.Registry,
	client MetricsClient,
) {
	tunnerTemperatureGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_tunner_temperature",
		Help: "Tunner temperature.",
	}, []string{})
	temperatureGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_temperature",
		Help: "Temperature.",
	}, []string{})
	operStateGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_box_oper_state",
		Help: "Operational state.",
	}, []string{})
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

	var data CMState
	err := client.GetMetrics(ctx, FnCMState, &data)
	if err != nil {
		log.Printf("Failed to get CMState: %v", err)
		return
	}

	tunnerTemperatureGauge.WithLabelValues().Set(float64(data.TunnerTemperature))
	temperatureGauge.WithLabelValues().Set(float64(data.Temperature))
	var val float64
	if data.OperState == OperStateOK {
		val = 1
	}
	operStateGauge.WithLabelValues().Set(val)
	wanIPv4AddrGauge.WithLabelValues(data.WANIPv4Addr).Set(1)
	for _, addr := range data.WANIPv6Addrs {
		wanIPv6AddrGauge.WithLabelValues(addr).Set(1)
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
