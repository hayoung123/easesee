#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const https = require('https');
const { spawn } = require('child_process');

const pkg = require('../package.json');
const VERSION = pkg.version;
const cacheDir = path.join(__dirname, '..', 'native');
const binPath = path.join(cacheDir, 'easesee');

const ASSET_MAP = {
  'darwin-arm64': 'easesee-darwin-arm64',
  'darwin-x64':   'easesee-darwin-amd64',
  'linux-arm64':  'easesee-linux-arm64',
  'linux-x64':    'easesee-linux-amd64',
};

function platformAsset() {
  const key = `${process.platform}-${process.arch}`;
  return { key, asset: ASSET_MAP[key] };
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const req = https.get(
      url,
      { headers: { 'User-Agent': 'easesee-installer' } },
      (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          res.resume();
          return download(res.headers.location, dest).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          res.resume();
          return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
        }
        const file = fs.createWriteStream(dest);
        res.pipe(file);
        file.on('finish', () => file.close((err) => err ? reject(err) : resolve()));
        file.on('error', reject);
      },
    );
    req.on('error', reject);
  });
}

async function ensureBinary() {
  if (fs.existsSync(binPath)) return;
  const { key, asset } = platformAsset();
  if (!asset) {
    console.error(`easesee: unsupported platform ${key}`);
    console.error('Supported: darwin-arm64, darwin-x64, linux-arm64, linux-x64');
    process.exit(1);
  }
  fs.mkdirSync(cacheDir, { recursive: true });
  const url = `https://github.com/hayoung123/easesee/releases/download/v${VERSION}/${asset}`;
  process.stderr.write(`easesee: downloading ${asset} (v${VERSION})...\n`);
  await download(url, binPath);
  fs.chmodSync(binPath, 0o755);
}

(async () => {
  try {
    await ensureBinary();
    const child = spawn(binPath, process.argv.slice(2), { stdio: 'inherit' });
    child.on('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code || 0);
      }
    });
    child.on('error', (err) => {
      console.error('easesee:', err.message);
      process.exit(1);
    });
  } catch (e) {
    console.error('easesee:', e.message);
    process.exit(1);
  }
})();
