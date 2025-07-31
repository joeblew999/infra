#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const pkg = require('./package.json');
const BINARY_NAME = Object.keys(pkg.bin)[0];

function getBinaryPath() {
  const platform = process.platform;
  const ext = platform === 'win32' ? '.exe' : '';
  const binaryPath = path.join(__dirname, 'bin', `${BINARY_NAME}${ext}`);
  
  if (!fs.existsSync(binaryPath)) {
    console.error(`❌ Binary not found at ${binaryPath}`);
    console.error('Try running: npm install');
    process.exit(1);
  }
  
  return binaryPath;
}

function runBinary() {
  const binaryPath = getBinaryPath();
  const args = process.argv.slice(2);
  
  // Spawn the binary with the same stdio streams
  const child = spawn(binaryPath, args, {
    stdio: 'inherit',
    cwd: process.cwd(),
    env: process.env
  });
  
  // Forward exit code
  child.on('exit', (code) => {
    process.exit(code || 0);
  });
  
  // Handle process termination
  process.on('SIGINT', () => {
    child.kill('SIGINT');
  });
  
  process.on('SIGTERM', () => {
    child.kill('SIGTERM');
  });
}

// Export for programmatic use
module.exports = {
  getBinaryPath,
  runBinary
};

// Run if called directly
if (require.main === module) {
  runBinary();
}