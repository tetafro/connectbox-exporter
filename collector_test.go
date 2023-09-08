package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector(map[string]MetricsClient{"test": &ConnectBox{}})
	require.Len(t, c.targets, 1)
}
