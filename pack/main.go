package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const (
	npmTokenEnvVar              = "NPM_TOKEN"
	npmUsernameEnvVar           = "NPM_USERNAME"
	npmTokenSettingsURLTemplate = "https://www.npmjs.com/settings/%s/tokens/"
	npmTokenSettingsURLGeneric  = "https://www.npmjs.com/settings/tokens/"
)

// Configuration for package.json
// Hardcoded for simplicity, can be made dynamic if needed, like using git reflection.

const (
	binaryName        = "infra"
	packageName       = "@joeblew99/infra"
	packageVersion    = "1.0.0"
	repositoryURL     = "https://github.com/joeblew99/infra.git"
	packageDescription = "A CLI tool for managing your Go project's npm package."
)

var buildMatrix = []struct {
	OS   string
	Arch string
}{
	{"windows", "amd64"},
	{"windows", "arm64"},
	{"darwin", "amd664"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"linux", "arm"},
}

var rootCmd = &cobra.Command{
	Use:   binaryName,
	Short: packageDescription,
	Long:  `A command-line interface for building, testing, and publishing your Go project as an npm package.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior if no subcommand is given
		cmd.Help()
	},
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publishes the npm package to the npm registry.",
	Long:  `This command builds the Go binary and then triggers the bun publish process.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running bun publish...")

		username := os.Getenv(npmUsernameEnvVar)
		var npmTokenSettingsURL string
		if username != "" {
			npmTokenSettingsURL = fmt.Sprintf(npmTokenSettingsURLTemplate, username)
		} else {
			npmTokenSettingsURL = npmTokenSettingsURLGeneric // Generic URL if username is not set
			log.Printf("Warning: %s environment variable is not set. The provided URL might not be specific to your account.\n", npmUsernameEnvVar)
		}

		// Check for NPM_TOKEN environment variable
		if os.Getenv(npmTokenEnvVar) == "" {
			log.Fatalf("Error: %s environment variable is not set. Please set it for bun publish to work. You can generate one at %s\n", npmTokenEnvVar, npmTokenSettingsURL)
		}

		// Use bun for publishing
		bunCmd := exec.Command("bun", "publish")
		bunCmd.Stdout = os.Stdout
		bunCmd.Stderr = os.Stderr
		err := bunCmd.Run()
		if err != nil {
			log.Fatalf("bun publish failed: %v\n", err)
		}
		fmt.Println("bun publish completed.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version [patch|minor|major|...]",
	Short: "Manages the npm package version.",
	Long:  `This command executes 'bun version' with the provided arguments. It updates the version in package.json and creates a Git tag.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running bun version...")

		bunArgs := append([]string{"version"}, args...)
		bunCmd := exec.Command("bun", bunArgs...)
		bunCmd.Stdout = os.Stdout
		bunCmd.Stderr = os.Stderr
		err := bunCmd.Run()
		if err != nil {
			log.Fatalf("bun version failed: %v\n", err)
		}
		fmt.Println("bun version completed.")
	},
}

// Removed packCmd as 'bun pack' is not a direct command

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the Go binary for the classic matrix of OS and architectures.",
	Long:  `This command compiles the main.go file into binaries for Windows, Darwin, and Linux across amd64 and arm64 (and arm for Linux).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building Go binaries...")
		for _, target := range buildMatrix {
			fmt.Printf("Building Go binary for OS: %s, Arch: %s...\n", target.OS, target.Arch)

			outputFileName := fmt.Sprintf("%s-%s-%s", binaryName, target.OS, target.Arch)
			if target.OS == "windows" {
				outputFileName += ".exe"
			}

			buildCmd := exec.Command("go", "build", "-o", outputFileName, "main.go")
			buildCmd.Env = os.Environ()
			buildCmd.Env = append(buildCmd.Env, fmt.Sprintf("GOOS=%s", target.OS))
			buildCmd.Env = append(buildCmd.Env, fmt.Sprintf("GOARCH=%s", target.Arch))

			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			err := buildCmd.Run()
			if err != nil {
				log.Fatalf("Go build for %s/%s failed: %v\n", target.OS, target.Arch, err)
			}
			fmt.Printf("Go build for %s/%s completed.\n", target.OS, target.Arch)
		}
		fmt.Println("All Go binaries built.")
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Removes build artifacts and node_modules.",
	Long:  `This command removes compiled Go binaries, temporary files, and the node_modules directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning up artifacts...")

		// Remove compiled Go binaries
		_ = os.RemoveAll(binaryName + "-*")  // Remove all binaries matching pattern
		_ = os.RemoveAll("bin")          // Remove the bin directory
		_ = os.RemoveAll("node_modules") // Remove node_modules

		fmt.Println("Cleanup completed.")
	},
}

var deprecateCmd = &cobra.Command{
	Use:   "deprecate <package-name>[@<version>] <message>",
	Short: "Deprecates a package or version on the npm registry.",
	Long:  `This command executes 'bun deprecate' to mark a package or specific version as deprecated.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatalf("Usage: %s deprecate <package-name>[@<version>] <message>\n", rootCmd.Use)
		}
		fmt.Println("Running bun deprecate...")
		bunArgs := append([]string{"deprecate"}, args...)
		bunCmd := exec.Command("bun", bunArgs...)
		bunCmd.Stdout = os.Stdout
		bunCmd.Stderr = os.Stderr
		err := bunCmd.Run()
		if err != nil {
			log.Fatalf("bun deprecate failed: %v\n", err)
		}
		fmt.Println("bun deprecate completed.")
	},
}

var unpublishCmd = &cobra.Command{
	Use:   "unpublish <package-name>[@<version>]",
	Short: "Removes a package or version from the npm registry.",
	Long:  `This command executes 'npm unpublish' to remove a package or specific version from the npm registry. Use with caution!`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("Usage: %s unpublish <package-name>[@<version>]\n", rootCmd.Use)
		}
		fmt.Println("Running npm unpublish...")
		npmArgs := append([]string{"unpublish"}, args...)
		npmCmd := exec.Command("npm", npmArgs...)
		npmCmd.Stdout = os.Stdout
		npmCmd.Stderr = os.Stderr
		err := npmCmd.Run()
		if err != nil {
			log.Fatalf("npm unpublish failed: %v\n", err)
		}
		fmt.Println("npm unpublish completed.")
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Automates GitHub release creation and asset upload.",
	Long:  `This command will build binaries for all supported platforms, create a GitHub release, and upload the binaries as assets.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running release process...")

		// 1. Clean old builds
		cleanCmd.Run(cmd, args)

		// 2. Build binaries
		buildCmd.Run(cmd, args)

		// 3. Create GitHub Release using gh CLI
		fmt.Println("Creating GitHub release...")
		tag := "v" + packageVersion
		
		// Find all built binaries
		files, err := os.ReadDir(".")
		if err != nil {
			log.Fatalf("Failed to read current directory: %v", err)
		}

		var assetArgs []string
		for _, file := range files {
			if strings.HasPrefix(file.Name(), binaryName + "-") && !file.IsDir() {
				assetArgs = append(assetArgs, file.Name())
			}
		}

		if len(assetArgs) == 0 {
			log.Fatalf("No binaries found to upload. Please ensure build command ran successfully.")
		}

		ghArgs := []string{"release", "create", tag, "--generate-notes"}
		ghArgs = append(ghArgs, assetArgs...)

		ghCmd := exec.Command("gh", ghArgs...)
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr
		err = ghCmd.Run()
		if err != nil {
			log.Fatalf("gh release create failed: %v\n", err)
		}
		fmt.Println("GitHub release created successfully!")
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates package.json from the template.",
	Long:  `This command reads package.json.template, performs variable substitutions, and writes the output to package.json.`,
	Run: func(cmd *cobra.Command, args []string) {
		// IMPORTANT: package.json.template is required for this command to function.
		// Do not delete package.json.template.
		fmt.Println("Generating package.json from template...")

		templateBytes, err := os.ReadFile("package.json.template")
		if err != nil {
			log.Fatalf("Failed to read package.json.template: %v", err)
		}

		tmpl, err := template.New("package.json").Parse(string(templateBytes))
		if err != nil {
			log.Fatalf("Failed to parse package.json.template: %v", err)
		}

		// Derive bugsURL and homepageURL from repositoryURL
		bugsURL := strings.TrimSuffix(repositoryURL, ".git") + "/issues"
		homepageURL := strings.TrimSuffix(repositoryURL, ".git") + "#readme"

		data := struct {
			Name        string
			Version     string
			RepoURL     string
			BugsURL     string
			HomepageURL string
			BinName     string
			Description string
		}{
			Name:        packageName,
			Version:     packageVersion,
			RepoURL:     repositoryURL,
			BugsURL:     bugsURL,
			HomepageURL: homepageURL,
			BinName:     binaryName,
			Description: packageDescription,
		}

		var processedTemplate bytes.Buffer
		err = tmpl.Execute(&processedTemplate, data)
		if err != nil {
			log.Fatalf("Failed to execute template: %v", err)
		}

		err = os.WriteFile("package.json", processedTemplate.Bytes(), 0644)
		if err != nil {
			log.Fatalf("Failed to write package.json: %v", err)
		}

		fmt.Println("Successfully generated package.json.")
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Displays package information.",
	Long:  `This command displays the package name, version, and repository URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Package Name: %s\n", packageName)
		fmt.Printf("Package Version: %s\n", packageVersion)
		fmt.Printf("Repository URL: %s\n", repositoryURL)
		fmt.Printf("Binary Name: %s\n", binaryName)
		fmt.Printf("Description: %s\n", packageDescription)
	},
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Runs all release-related commands in sequence (clean, build, pack, publish, release).",
	Long:  `This command orchestrates the full release process: cleaning artifacts, building binaries for all platforms, creating the npm package, publishing to npm, and creating a GitHub release.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running all release commands...")

		// Clean
		cleanCmd.Run(cmd, args)

		// Build
		buildCmd.Run(cmd, args)

		// Pack (Implicitly handled by bun publish, or manually if needed)
		// packCmd.Run(cmd, args) // Removed as 'bun pack' is not a direct command

		// Publish
		publishCmd.Run(cmd, args)
		// Release
		releaseCmd.Run(cmd, args)

		fmt.Println("All release commands completed.")
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(versionCmd)
	// rootCmd.AddCommand(packCmd) // Removed as 'bun pack' is not a direct command
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(deprecateCmd)
	rootCmd.AddCommand(unpublishCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(allCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(infoCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
