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

	// CPU assertions
	assert.GreaterOrEqual(t, metrics.CPU.Percent, 0.0)

	// Memory assertions
	assert.Greater(t, metrics.Memory.TotalMB, uint64(0))
	assert.GreaterOrEqual(t, metrics.Memory.UsedMB, uint64(0))
	assert.GreaterOrEqual(t, metrics.Memory.UsedPercent, 0.0)

	// Disk assertions
	assert.NotEmpty(t, metrics.Disk.MountPoints)
	rootDisk, ok := metrics.Disk.MountPoints["/"]
	assert.True(t, ok, "Disk usage for root (/) should be present")
	assert.Greater(t, rootDisk.TotalGB, uint64(0))
	assert.GreaterOrEqual(t, rootDisk.UsedGB, uint64(0))
	assert.GreaterOrEqual(t, rootDisk.UsedPercent, 0.0)
}
