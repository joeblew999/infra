package cmd

import "testing"

func TestGenerateDemo(t *testing.T) {
	err := GenerateDemo()
	if err != nil {
		t.Fatalf("GenerateDemo failed: %v", err)
	}
}