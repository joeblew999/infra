package config

import (
	"fmt"
	"os"
)

var requiredEnvVars = []string{
	EnvVarFlyAppName,
	"FLY_API_TOKEN",
}

// ValidateEnvironment checks if all required environment variables are set
func ValidateEnvironment() error {
	missing := []string{}
	
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}
	
	return nil
}

// GetMissingEnvVars returns missing required environment variables
func GetMissingEnvVars() []string {
	missing := []string{}
	
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}
	
	return missing
}

// GetEnvStatus returns status of all expected environment variables
func GetEnvStatus() map[string]string {
	allVars := []string{
		EnvVarEnvironment,
		EnvVarFlyAppName,
		EnvVarKoDockerRepo,
		"FLY_API_TOKEN",
		"FLY_REGION",
	}
	
	status := make(map[string]string)
	for _, envVar := range allVars {
		value := os.Getenv(envVar)
		if value == "" {
			status[envVar] = "not set"
		} else {
			status[envVar] = "set"
		}
	}
	
	return status
}