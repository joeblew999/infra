#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function cleanupBinaries() {
  const binDir = path.join(__dirname, 'bin');
  
  if (fs.existsSync(binDir)) {
    try {
      fs.rmSync(binDir, { recursive: true, force: true });
      console.log('🧹 Cleaned up downloaded binaries');
    } catch (error) {
      console.warn('Warning: Could not clean up binaries:', error.message);
    }
  }
}

// Run cleanup
cleanupBinaries();