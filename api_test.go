package main

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
)

// <cm_network_access>Allowed</cm_network_access></cm_system_info>
func TestCMSystemInfo_UnmarshalXML(t *testing.T) {
	t.Run("valid xml", func(t *testing.T) {
		data := `<?xml version="1.0" encoding="utf-8"?>` +
			`<cmsysteminfo>` +
			`<cm_docsis_mode>DOCSIS 3.0</cm_docsis_mode>` +
			`<cm_hardware_version>5.01</cm_hardware_version>` +
			`<cm_mac_addr>00:00:00:00:00:00</cm_mac_addr>` +
			`<cm_serial_number>AAAAAAAAAAAA</cm_serial_number>` +
			`<cm_system_uptime>10day(s)20h:15m:30s</cm_system_uptime>` +
			`<cm_network_access>Allowed</cm_network_access>` +
			`</cmsysteminfo>`

		var cminfo CMSystemInfo
		err := xml.Unmarshal([]byte(data), &cminfo)
		require.NoError(t, err)

		expected := CMSystemInfo{
			DocsisMode:      "DOCSIS 3.0",
			HardwareVersion: "5.01",
			MacAddr:         "00:00:00:00:00:00",
			SerialNumber:    "AAAAAAAAAAAA",
			SystemUptime:    936930,
			NetworkAccess:   "Allowed",
		}
		require.Equal(t, expected, cminfo)
	})

	t.Run("invalid duration", func(t *testing.T) {
		data := `<?xml version="1.0" encoding="utf-8"?>` +
			`<cmsysteminfo>` +
			`<cm_docsis_mode>DOCSIS 3.0</cm_docsis_mode>` +
			`<cm_hardware_version>5.01</cm_hardware_version>` +
			`<cm_mac_addr>00:00:00:00:00:00</cm_mac_addr>` +
			`<cm_serial_number>AAAAAAAAAAAA</cm_serial_number>` +
			`<cm_system_uptime>hello, world</cm_system_uptime>` +
			`<cm_network_access>Allowed</cm_network_access>` +
			`</cmsysteminfo>`

		var cminfo CMSystemInfo
		err := xml.Unmarshal([]byte(data), &cminfo)
		require.ErrorContains(t, err, "invalid duration string")
	})

	t.Run("invalid xml", func(t *testing.T) {
		data := `<?xml version="1.0" encoding="utf-8"?><cmsysteminfo>`

		var cminfo CMSystemInfo
		err := xml.Unmarshal([]byte(data), &cminfo)
		require.ErrorContains(t, err, "XML syntax error")
	})
}

func TestCMState_UnmarshalXML(t *testing.T) {
	t.Run("valid xml", func(t *testing.T) {
		data := `<?xml version="1.0" encoding="utf-8"?>` +
			`<cmstate>` +
			`<TunnerTemperature>80</TunnerTemperature>` +
			`<Temperature>60</Temperature>` +
			`<OperState>OPERATIONAL</OperState>` +
			`<wan_ipv4_addr>1.1.1.1</wan_ipv4_addr>` +
			`<wan_ipv6_addr>` +
			`<wan_ipv6_addr_entry>0000:0000:0000:0000:0000:0000:0000:0001/128</wan_ipv6_addr_entry>` +
			`<wan_ipv6_addr_entry>0000:0000:0000:0000:0000:0000:0000:0002/128</wan_ipv6_addr_entry>` +
			`</wan_ipv6_addr>` +
			`</cmstate>`

		var cmstate CMState
		err := xml.Unmarshal([]byte(data), &cmstate)
		require.NoError(t, err)

		expected := CMState{
			TunnerTemperature: 26,
			Temperature:       15,
			OperState:         "OPERATIONAL",
			WANIPv4Addr:       "1.1.1.1",
			WANIPv6Addrs: []string{
				"0000:0000:0000:0000:0000:0000:0000:0001/128",
				"0000:0000:0000:0000:0000:0000:0000:0002/128",
			},
		}
		require.Equal(t, expected, cmstate)
	})

	t.Run("invalid xml", func(t *testing.T) {
		data := `<?xml version="1.0" encoding="utf-8"?><cmstate>`

		var cmstate CMState
		err := xml.Unmarshal([]byte(data), &cmstate)
		require.ErrorContains(t, err, "XML syntax error")
	})
}

func TestFahrenheitToCelsius(t *testing.T) {
	testCases := []struct {
		name       string
		fahrenheit int
		celsius    int
	}{
		{
			name:       "above zero",
			fahrenheit: 50,
			celsius:    10,
		},
		{
			name:       "below zero",
			fahrenheit: -50,
			celsius:    -45,
		},
		{
			name:       "zero fahrenheit",
			fahrenheit: 0,
			celsius:    -17,
		},
		{
			name:       "zero celsius",
			fahrenheit: 32,
			celsius:    0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := fahrenheitToCelsius(tc.fahrenheit)
			if c != tc.celsius {
				t.Fatalf("Wrong column\n  expected: %d\n       got: %d", tc.celsius, c)
			}
		})
	}
}
