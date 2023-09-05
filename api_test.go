package main

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"
)

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
