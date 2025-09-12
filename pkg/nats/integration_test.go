package nats_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/stretchr/testify/require"
)

// setupS3IntegrationTest starts the NATS server and S3 gateway for integration tests.
// It returns the S3 endpoint URL, the NATS server address, and a cleanup function.
func setupS3IntegrationTest(t *testing.T) (string, string, func()) {
	t.Helper()

	// Start embedded NATS server
	natsAddr, natsCleanup, err := nats.StartEmbeddedNATS(context.Background())
	require.NoError(t, err)

	// Start the nats-s3 gateway
	nats.StartS3GatewaySupervised(natsAddr)

	// Give the gateway a moment to start
	time.Sleep(1 * time.Second)

	endpoint := fmt.Sprintf("http://localhost:%s", config.GetNatsS3Port())

	cleanup := func() {
		goreman.StopAll()
		natsCleanup()
	}

	return endpoint, natsAddr, cleanup
}

func TestS3GatewayIntegration_BlackBox(t *testing.T) {
	endpoint, _, cleanup := setupS3IntegrationTest(t) // natsAddr not needed here
	defer cleanup()

	// Test data
	bucketName := "test-bucket-blackbox"
	objectKey := "test-object"
	fileContent := "Hello, NATS S3!"
	bucketURL := fmt.Sprintf("%s/%s", endpoint, bucketName)
	objectURL := fmt.Sprintf("%s/%s", bucketURL, objectKey)

	// 1. Create bucket
	req, err := http.NewRequest(http.MethodPut, bucketURL, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "failed to create bucket")

	// 2. Upload file
	req, err = http.NewRequest(http.MethodPut, objectURL, bytes.NewReader([]byte(fileContent)))
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "failed to upload file")

	// 3. Download file
	resp, err = http.Get(objectURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "failed to download file")

	// 4. Verify content
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, fileContent, string(body), "downloaded content does not match")

	// 5. Delete file
	req, err = http.NewRequest(http.MethodDelete, objectURL, nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "failed to delete file")

	// 6. Delete bucket
	req, err = http.NewRequest(http.MethodDelete, bucketURL, nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode, "failed to delete bucket")
}

func TestS3GatewayIntegration_WhiteBox(t *testing.T) {
	endpoint, natsAddr, cleanup := setupS3IntegrationTest(t)
	defer cleanup()

	// Test data
	bucketName := "test-bucket-whitebox"
	objectKey := "test-object.txt"
	fileContent := "Hello, NATS CLI!"
	bucketURL := fmt.Sprintf("%s/%s", endpoint, bucketName)
	objectURL := fmt.Sprintf("%s/%s", bucketURL, objectKey)

	// 1. Create bucket and upload file via S3 API
	req, err := http.NewRequest(http.MethodPut, bucketURL, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodPut, objectURL, bytes.NewReader([]byte(fileContent)))
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. Verify with NATS CLI
	natsCliPath, err := config.GetAbsoluteDepPath("nats")
	require.NoError(t, err)

	// Verify object exists in bucket
	cmd := exec.Command(natsCliPath, "obj", "ls", bucketName, "--server", natsAddr)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "nats obj ls failed: %s", string(output))
	require.Contains(t, string(output), objectKey, "object key not found in nats obj ls output")

	// Verify object info
	cmd = exec.Command(natsCliPath, "obj", "info", bucketName, objectKey, "--server", natsAddr)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "nats obj info failed: %s", string(output))
	require.Contains(t, string(output), fmt.Sprintf("Size: %d", len(fileContent)), "file size not correct in nats obj info output")

	// 3. Cleanup via S3 API
	req, err = http.NewRequest(http.MethodDelete, objectURL, nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	req, err = http.NewRequest(http.MethodDelete, bucketURL, nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
}
