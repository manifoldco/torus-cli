#!/usr/bin/env node
var fs = require('fs');
var path = require('path');

if (['linux', 'darwin'].indexOf(process.platform) === -1) {
  console.log('torus only supports linux and darwin amd64');
  throw new Error('Incompatible Platform: ' + process.platform);
}

var os = process.platform;
var arch = 'amd64';

var torusPath = path.join(__dirname, '../bin/torus');

if (fs.existsSync(torusPath)) {
  fs.unlinkSync(torusPath);
}

fs.symlinkSync('torus-' + os + '-' + arch, torusPath);
