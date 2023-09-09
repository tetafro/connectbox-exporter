package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector(map[string]MetricsClient{"test": &ConnectBox{}})
	require.Len(t, c.targets, 1)
}

func TestCollector_ServeHTTP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		metrics := NewMockMetricsClient(ctrl)

		metrics.EXPECT().Login(gomock.Any()).Return(nil)

		var cmSystemInfoData CMSystemInfo
		metrics.EXPECT().GetMetrics(
			gomock.Any(), FnCMSystemInfo, &cmSystemInfoData,
		).Do(func(ctx context.Context, fn string, out any) error {
			data := out.(*CMSystemInfo)
			data.DocsisMode = "DocsisMode"
			data.HardwareVersion = "HardwareVersion"
			data.MacAddr = "MacAddr"
			data.SerialNumber = "SerialNumber"
			data.SystemUptime = 100
			data.NetworkAccess = NetworkAccessAllowed
			return nil
		})

		var lanUserTableData LANUserTable
		metrics.EXPECT().GetMetrics(
			gomock.Any(), FnLANUserTable, &lanUserTableData,
		).Do(func(ctx context.Context, fn string, out any) error {
			data := out.(*LANUserTable)
			data.Ethernet = []LANUserTableClientInfo{{
				Interface:   "EthernetInterface",
				IPv4Addr:    "EthernetIPv4Addr",
				Index:       "EthernetIndex",
				InterfaceID: "EthernetInterfaceID",
				Hostname:    "EthernetHostname",
				MACAddr:     "EthernetMACAddr",
				Method:      "EthernetMethod",
				LeaseTime:   "EthernetLeaseTime",
				Speed:       "EthernetSpeed",
			}}
			data.WIFI = []LANUserTableClientInfo{{
				Interface:   "WIFIInterface",
				IPv4Addr:    "WIFIIPv4Addr",
				Index:       "WIFIIndex",
				InterfaceID: "WIFIInterfaceID",
				Hostname:    "WIFIHostname",
				MACAddr:     "WIFIMACAddr",
				Method:      "WIFIMethod",
				LeaseTime:   "WIFILeaseTime",
				Speed:       "WIFISpeed",
			}}
			return nil
		})

		var cmStateData CMState
		metrics.EXPECT().GetMetrics(
			gomock.Any(), FnCMState, &cmStateData,
		).Do(func(ctx context.Context, fn string, out any) error {
			data := out.(*CMState)
			data.TunnerTemperature = 10
			data.Temperature = 20
			data.OperState = OperStateOK
			data.WANIPv4Addr = "WANIPv4Addr"
			data.WANIPv6Addrs = []string{"WANIPv6Addr"}
			return nil
		})

		metrics.EXPECT().Logout(gomock.Any()).Return(nil)

		col := &Collector{
			targets: map[string]MetricsClient{
				"127.0.0.1": metrics,
			},
		}

		req, err := http.NewRequest(http.MethodGet, "/probe?target=127.0.0.1", nil)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		col.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		want := strings.Join([]string{
			`# HELP connect_box_cm_docsis_mode DocSis mode.`,
			`# TYPE connect_box_cm_docsis_mode gauge`,
			`connect_box_cm_docsis_mode{mode="DocsisMode"} 1`,
			`# HELP connect_box_cm_hardware_version Hardware version.`,
			`# TYPE connect_box_cm_hardware_version gauge`,
			`connect_box_cm_hardware_version{version="HardwareVersion"} 1`,
			`# HELP connect_box_cm_mac_addr MAC address.`,
			`# TYPE connect_box_cm_mac_addr gauge`,
			`connect_box_cm_mac_addr{addr="MacAddr"} 1`,
			`# HELP connect_box_cm_network_access Network access.`,
			`# TYPE connect_box_cm_network_access gauge`,
			`connect_box_cm_network_access 1`,
			`# HELP connect_box_cm_serial_number Serial number.`,
			`# TYPE connect_box_cm_serial_number gauge`,
			`connect_box_cm_serial_number{sn="SerialNumber"} 1`,
			`# HELP connect_box_cm_system_uptime System uptime.`,
			`# TYPE connect_box_cm_system_uptime gauge`,
			`connect_box_cm_system_uptime 100`,
			`# HELP connect_box_lan_client LAN client.`,
			`# TYPE connect_box_lan_client gauge`,
			`connect_box_lan_client{` +
				`MACAddr="EthernetMACAddr",connection="ethernet",` +
				`hostname="EthernetHostname",interface="EthernetInterface",` +
				`ipv4_addr="EthernetIPv4Addr"} 1`,
			`connect_box_lan_client{MACAddr="WIFIMACAddr",connection="wifi",` +
				`hostname="WIFIHostname",interface="WIFIInterface",` +
				`ipv4_addr="WIFIIPv4Addr"} 1`,
			`# HELP connect_box_oper_state Operational state.`,
			`# TYPE connect_box_oper_state gauge`,
			`connect_box_oper_state 1`,
			`# HELP connect_box_temperature Temperature.`,
			`# TYPE connect_box_temperature gauge`,
			`connect_box_temperature 20`,
			`# HELP connect_box_tunner_temperature Tunner temperature.`,
			`# TYPE connect_box_tunner_temperature gauge`,
			`connect_box_tunner_temperature 10`,
			`# HELP connect_box_wan_ipv4_addr WAN IPv4 address.`,
			`# TYPE connect_box_wan_ipv4_addr gauge`,
			`connect_box_wan_ipv4_addr{ip="WANIPv4Addr"} 1`,
			`# HELP connect_box_wan_ipv6_addr WAN IPv6 address.`,
			`# TYPE connect_box_wan_ipv6_addr gauge`,
			`connect_box_wan_ipv6_addr{ip="WANIPv6Addr"} 1`,
		}, "\n") + "\n"
		require.Equal(t, want, rec.Body.String())
	})

	t.Run("no target", func(t *testing.T) {
		col := &Collector{
			targets: map[string]MetricsClient{},
		}

		req, err := http.NewRequest(http.MethodGet, "/probe?target=127.0.0.1", nil)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		col.ServeHTTP(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("failed to login", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		metrics := NewMockMetricsClient(ctrl)
		metrics.EXPECT().Login(gomock.Any()).Return(errors.New("fail"))

		col := &Collector{
			targets: map[string]MetricsClient{
				"127.0.0.1": metrics,
			},
		}

		req, err := http.NewRequest(http.MethodGet, "/probe?target=127.0.0.1", nil)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		col.ServeHTTP(rec, req)

		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
