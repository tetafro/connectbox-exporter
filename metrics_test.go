package main

import "testing"

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
