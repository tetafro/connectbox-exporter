package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector(map[string]*ConnectBox{"test": {}})
	require.Len(t, c.targets, 1)
}
