# Your Go Tool - NPM Package

A Node.js/Deno/Bun wrapper for your Go binary, making it easy to install and use across JavaScript runtimes.

## Installation

### Node.js
```bash
npm install your-go-tool
# or globally
npm install -g your-go-tool
```

### Deno
```bash
# Using npm compatibility
deno run --allow-run npm:your-go-tool
```

### Bun
```bash
bun install your-go-tool
# or globally  
bun install -g your-go-tool
```

## Usage

### Command Line
```bash
# If installed locally
npx your-tool [options]

# If installed globally
your-tool [options]

# With Deno
deno run --allow-run npm:your-go-tool [options]

# With Bun
bunx your-tool [options]
```

### Programmatic Usage (Node.js)
```javascript
const { spawn } = require('child_process');
const { getBinaryPath } = require('your-go-tool');

// Get the path to the binary
const binaryPath = getBinaryPath();

// Run with custom arguments
const child = spawn(binaryPath, ['--help'], { stdio: 'inherit' });
```

## Supported Platforms

- **macOS**: x64, ARM64 (Apple Silicon)
- **Linux**: x64, ARM64, ARM
- **Windows**: x64

## How It Works

1. During `npm install`, the package automatically downloads the correct binary for your platform from GitHub releases
2. The binary is stored in the package's `bin/` directory
3. The wrapper script (`index.js`) forwards all arguments to the Go binary
4. Exit codes and signals are properly handled

## Development

### Building Your Go Binary for Release

Make sure your Go project builds binaries for all supported platforms:

```bash
# Build for different platforms
GOOS=darwin GOARCH=amd64 go build -o your-tool-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o your-tool-darwin-arm64
GOOS=linux GOARCH=amd64 go build -o your-tool-linux-amd64
GOOS=linux GOARCH=arm64 go build -o your-tool-linux-arm64
GOOS=linux GOARCH=arm go build -o your-tool-linux-arm
GOOS=windows GOARCH=amd64 go build -o your-tool-windows-amd64.exe
```

### GitHub Release Setup

1. Create a new release on GitHub with version tag (e.g., `v1.0.0`)
2. Upload all platform binaries as release assets
3. Update the version in `package.json` to match

### Publishing to NPM

```bash
npm publish
```

## Troubleshooting

### Binary Not Found
If you see "Binary not found" errors:
```bash
npm install  # Re-run installation
```

### Platform Not Supported
Check that your platform is in the supported list above. The package will show available platforms if yours isn't supported.

### Download Issues
- Check your internet connection
- Verify the GitHub release exists
- Ensure all required binaries are uploaded to the release

## License

MIT