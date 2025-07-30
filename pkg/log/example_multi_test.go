package log

import (
	"testing"
)

func ExampleInitMultiLogger() {
	// Example configuration for multi-destination logging
	config := MultiConfig{
		Destinations: []DestinationConfig{
			{
				Type:   "stdout",
				Level:  "info",
				Format: "json",
			},
			{
				Type:   "file",
				Level:  "debug",
				Format: "text",
				Path:   "/tmp/app.log",
			},
			{
				Type:    "file",
				Level:   "error",
				Format:  "json",
				Path:    "/tmp/app-errors.log",
			},
		},
	}

	// Initialize multi-destination logging
	if err := InitMultiLogger(config); err != nil {
		panic(err)
	}

	// Now all logs will go to multiple destinations
	Info("Application started", "version", "1.0.0")
	Debug("Debug information", "details", "some debug data")
	Error("Something went wrong", "error", "timeout")
}

func TestInitMultiLogger(t *testing.T) {
	config := MultiConfig{
		Destinations: []DestinationConfig{
			{
				Type:   "stdout",
				Level:  "debug",
				Format: "text",
			},
		},
	}

	if err := InitMultiLogger(config); err != nil {
		t.Errorf("InitMultiLogger failed: %v", err)
	}

	// Test logging
	Info("Test message", "test", true)
}