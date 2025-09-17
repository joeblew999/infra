package gops_test

import (
	"testing"

	"github.com/joeblew999/infra/pkg/gops"
	"github.com/stretchr/testify/assert"
)

func TestGetSystemMetrics(t *testing.T) {
	metrics, err := gops.GetSystemMetrics()
	assert.NoError(t, err)

	assert.NotEmpty(t, metrics.ServerID)
	assert.NotEmpty(t, metrics.Timestamp)

	// Runtime assertions
	assert.GreaterOrEqual(t, metrics.Runtime.NumGoroutines, 1)
	assert.Greater(t, metrics.Runtime.NumCPU, 0)
	assert.GreaterOrEqual(t, metrics.Runtime.MemAlloc, uint64(0))
	assert.GreaterOrEqual(t, metrics.Runtime.MemTotal, uint64(0))
	assert.GreaterOrEqual(t, metrics.Runtime.MemSys, uint64(0))
	assert.GreaterOrEqual(t, metrics.Runtime.NumGC, uint32(0))
	assert.NotEmpty(t, metrics.Runtime.GOOS)
	assert.NotEmpty(t, metrics.Runtime.GOARCH)
}
