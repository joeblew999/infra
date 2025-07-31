#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { exec } = require('child_process');
const { pipeline } = require('stream');
const { promisify } = require('util');

const pipelineAsync = promisify(pipeline);

// Read configuration from package.json
const pkg = require('./package.json');
const BINARY_NAME = pkg.name.split('/').pop(); // a scoped package name like @user/repo becomes 'repo'
const GITHUB_REPO = new URL(pkg.repository.url).pathname.substring(1).replace('.git', '');
const VERSION = pkg.version;

// Platform mappings
const PLATFORM_MAPPING = {
  'darwin': 'darwin',
  'linux': 'linux',
  'win32': 'windows'
};

const ARCH_MAPPING = {
  'x64': 'amd64',
  'arm64': 'arm64',
  'arm': 'arm'
};

function getPlatformInfo() {
  const platform = PLATFORM_MAPPING[process.platform];
  const arch = ARCH_MAPPING[process.arch];
  
  if (!platform || !arch) {
    throw new Error(`Unsupported platform: ${process.platform}-${process.arch}`);
  }
  
  return { platform, arch };
}

function getBinaryUrl(platform, arch) {
  const ext = platform === 'windows' ? '.exe' : '';
  const filename = `${BINARY_NAME}-${platform}-${arch}${ext}`;
  
  // GitHub releases URL format
  return `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${filename}`;
}

function getBinaryPath(platform) {
  const binDir = path.join(__dirname, 'bin');
  const ext = platform === 'windows' ? '.exe' : '';
  return path.join(binDir, `${BINARY_NAME}${ext}`);
}

async function downloadFile(url, destinationPath) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading ${url}...`);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        return downloadFile(response.headers.location, destinationPath)
          .then(resolve)
          .catch(reject);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Download failed with status ${response.statusCode}`));
        return;
      }
      
      const writeStream = fs.createWriteStream(destinationPath);
      
      pipelineAsync(response, writeStream)
        .then(() => {
          console.log(`Downloaded to ${destinationPath}`);
          resolve();
        })
        .catch(reject);
    }).on('error', reject);
  });
}

async function makeExecutable(filePath) {
  return new Promise((resolve, reject) => {
    exec(`chmod +x "${filePath}"`, (error) => {
      if (error) {
        reject(error);
      } else {
        resolve();
      }
    });
  });
}

async function install() {
  try {
    const { platform, arch } = getPlatformInfo();
    const url = getBinaryUrl(platform, arch);
    const binaryPath = getBinaryPath(platform);
    
    // Create bin directory if it doesn't exist
    const binDir = path.dirname(binaryPath);
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    // Download the binary
    await downloadFile(url, binaryPath);
    
    // Make it executable on Unix-like systems
    if (platform !== 'windows') {
      await makeExecutable(binaryPath);
    }
    
    console.log(`✅ ${BINARY_NAME} installed successfully!`);
    console.log(`Run with: npx ${BINARY_NAME} or ${BINARY_NAME} if installed globally`);
    
  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    
    // Provide helpful error messages
    if (error.message.includes('Unsupported platform')) {
      console.error('\nSupported platforms:');
      console.error('- macOS (darwin) - x64, arm64');
      console.error('- Linux - x64, arm64, arm');
      console.error('- Windows - x64');
    } else if (error.message.includes('Download failed')) {
      console.error('\nPlease check:');
      console.error('1. Internet connection');
      console.error('2. GitHub release exists for this version');
      console.error(`3. Binary exists: ${error.url || 'check release assets'}`);
    }
    
    process.exit(1);
  }
}

// Run installation
install();