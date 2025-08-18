package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadFile downloads a file from URL to destination path with progress tracking
func DownloadFile(url, destPath string, showProgress bool) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer out.Close()

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d for %s", resp.StatusCode, url)
	}

	// Download with or without progress
	fileName := filepath.Base(destPath)
	contentLength := resp.ContentLength

	if !showProgress {
		// Simple download without progress
		_, err = io.Copy(out, resp.Body)
		return err
	}

	if contentLength > 0 {
		// Download with full progress bar
		progressReader := NewProgressReader(resp.Body, contentLength, fileName)
		_, err = io.Copy(out, progressReader)
		progressReader.Finish()
	} else {
		// Download with simple progress (unknown size)
		progressWriter := NewSimpleProgressWriter(fileName)
		_, err = io.Copy(io.MultiWriter(out, progressWriter), resp.Body)
		progressWriter.Finish()
	}

	if err != nil {
		return fmt.Errorf("failed to copy downloaded content: %w", err)
	}

	return nil
}

// DownloadToTemp downloads a file to a temporary location with progress
func DownloadToTemp(url, prefix string, showProgress bool) (string, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFile.Close() // Close so DownloadFile can write to it

	// Download to temp file
	if err := DownloadFile(url, tempFile.Name(), showProgress); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}