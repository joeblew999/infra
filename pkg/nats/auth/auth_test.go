package auth_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/nats/auth"
	"github.com/stretchr/testify/require"
)

func TestEnsureArtifacts(t *testing.T) {
	t.Setenv("ENVIRONMENT", "test")

	store := config.GetNATSAuthStorePath()
	base := filepath.Dir(store)
	require.NoError(t, os.RemoveAll(base))
	t.Cleanup(func() {
		_ = os.RemoveAll(base)
	})

	ctx := context.Background()
	artifacts, err := auth.Ensure(ctx)
	require.NoError(t, err)
	require.NotNil(t, artifacts)
	require.NotEmpty(t, artifacts.OperatorJWT)
	require.NotEmpty(t, artifacts.SystemAccountID)
	require.NotEmpty(t, artifacts.ApplicationAccountID)

	for _, path := range []string{
		config.GetNATSOperatorJWTPath(),
		config.GetNATSAccountJWTPath(config.NATSSystemAccountName),
		config.GetNATSAccountJWTPath(config.NATSApplicationAccount),
		config.GetNATSApplicationCredsPath(),
		config.GetNATSSystemCredsPath(),
	} {
		_, err := os.Stat(path)
		require.NoErrorf(t, err, "expected auth material at %s", path)
	}

	// Run again to make sure it is idempotent
	artifactsAgain, err := auth.Ensure(ctx)
	require.NoError(t, err)
	require.Equal(t, artifacts.SystemAccountID, artifactsAgain.SystemAccountID)
	require.Equal(t, artifacts.ApplicationAccountID, artifactsAgain.ApplicationAccountID)
}
