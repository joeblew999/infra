package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// GetEnvOrPrompt gets environment variable with platform-native fallback chain:
// 1. Environment variable
// 2. OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
// 3. Interactive prompt with secure storage option
func GetEnvOrPrompt(key, description string) (string, error) {
	// Step 1: Check environment variable
	if value := os.Getenv(key); value != "" {
		return value, nil
	}

	// Step 2: Try platform-native keychain
	if value, err := getFromKeychain(key); err == nil && value != "" {
		return value, nil
	}

	// Step 3: Interactive prompt
	fmt.Printf("\n%s is required but not set.\n", key)
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	
	value, err := promptSecure(fmt.Sprintf("Enter %s: ", key))
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", key, err)
	}

	// Offer to store in keychain for future use
	if shouldStore := promptYesNo(fmt.Sprintf("Store %s in system keychain for future use?", key)); shouldStore {
		if err := SetEnvSecret(key, value); err != nil {
			fmt.Printf("Warning: Failed to store in keychain: %v\n", err)
		} else {
			fmt.Printf("✅ %s stored in system keychain\n", key)
		}
	}

	return value, nil
}

// SetEnvSecret stores value in platform-native keychain
func SetEnvSecret(key, value string) error {
	switch runtime.GOOS {
	case "darwin":
		return setMacOSKeychain(key, value)
	case "windows":
		return setWindowsCredential(key, value)
	case "linux":
		return setLinuxSecret(key, value)
	default:
		return fmt.Errorf("platform %s not supported for keychain storage", runtime.GOOS)
	}
}

// BootstrapRequiredEnvs validates and prompts for missing environment variables
func BootstrapRequiredEnvs() error {
	envDescriptions := map[string]string{
		EnvVarFlyAppName: "Fly.io application name for deployment target",
		"GITHUB_TOKEN": "GitHub Personal Access Token for container registry access (optional, can fallback to Fly registry)",
	}

	fmt.Println("🔧 Bootstrapping deployment environment...")
	
	// Get missing vars but exclude FLY_API_TOKEN (handled by flyctl auth flow)
	missing := []string{}
	for _, envVar := range []string{EnvVarFlyAppName} {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}
	
	// Check GITHUB_TOKEN separately as it's optional
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Printf("\nGITHUB_TOKEN (optional - can use Fly registry as fallback)\n")
		fmt.Printf("Description: %s\n", envDescriptions["GITHUB_TOKEN"])
		if promptYesNo("Set GITHUB_TOKEN now?") {
			missing = append(missing, "GITHUB_TOKEN")
		} else {
			fmt.Println("⏭️  Skipping GITHUB_TOKEN - will use Fly registry")
		}
	}
	
	if len(missing) == 0 {
		fmt.Println("✅ All required environment variables are set")
		fmt.Println("ℹ️  Fly.io authentication will be handled by flyctl during deployment")
		return nil
	}

	fmt.Printf("📋 Found %d missing environment variables\n", len(missing))

	for _, envVar := range missing {
		description := envDescriptions[envVar]
		
		value, err := GetEnvOrPrompt(envVar, description)
		if err != nil {
			return fmt.Errorf("failed to get %s: %w", envVar, err)
		}

		// Set in current process environment
		os.Setenv(envVar, value)
		fmt.Printf("✅ %s configured\n", envVar)
	}

	fmt.Println("🎉 Environment bootstrap completed successfully")
	fmt.Println("ℹ️  Fly.io authentication will be handled by flyctl during deployment")
	return nil
}

// Platform-specific implementations

func getFromKeychain(key string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacOSKeychain(key)
	case "windows":
		return getWindowsCredential(key)
	case "linux":
		return getLinuxSecret(key)
	default:
		return "", fmt.Errorf("platform not supported")
	}
}

// macOS Keychain implementation
func getMacOSKeychain(key string) (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "infra-cli", "-a", key, "-w")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func setMacOSKeychain(key, value string) error {
	// Delete existing entry first (ignore errors)
	exec.Command("security", "delete-generic-password", "-s", "infra-cli", "-a", key).Run()
	
	// Add new entry
	cmd := exec.Command("security", "add-generic-password", "-s", "infra-cli", "-a", key, "-w", value)
	return cmd.Run()
}

// Windows Credential Manager implementation
func getWindowsCredential(key string) (string, error) {
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("$cred = Get-StoredCredential -Target 'infra-cli:%s' -ErrorAction SilentlyContinue; if ($cred) { $cred.GetNetworkCredential().Password }", key))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func setWindowsCredential(key, value string) error {
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("New-StoredCredential -Target 'infra-cli:%s' -UserName '%s' -Password '%s' -Persist LocalMachine", key, key, value))
	return cmd.Run()
}

// Linux Secret Service implementation (fallback to simple approach)
func getLinuxSecret(key string) (string, error) {
	// Try secret-tool first (libsecret)
	cmd := exec.Command("secret-tool", "lookup", "application", "infra-cli", "key", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func setLinuxSecret(key, value string) error {
	cmd := exec.Command("secret-tool", "store", "--label", fmt.Sprintf("infra-cli %s", key), "application", "infra-cli", "key", key)
	cmd.Stdin = strings.NewReader(value)
	return cmd.Run()
}

// Helper functions

func promptSecure(prompt string) (string, error) {
	fmt.Print(prompt)
	
	// Check if stdin is a terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		// Non-interactive mode, read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Text(), nil
		}
		return "", fmt.Errorf("failed to read from stdin")
	}
	
	// Interactive mode, use hidden input for sensitive data
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // Add newline after hidden input
	
	return string(bytePassword), nil
}

func promptYesNo(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}
	return false
}