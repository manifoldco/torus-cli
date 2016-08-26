#!/usr/bin/env node
var fs = require('fs');
var path = require('path');

if (['linux', 'darwin'].indexOf(process.platform) === -1) {
  console.log('ag only supports linux and darwin amd64');
  throw new Error('Incompatible Platform: ' + process.platform);
}

var os = process.platform;
var arch = 'amd64';

var agPath = path.join(__dirname, '../bin/ag');

if (fs.existsSync(agPath)) {
  fs.unlinkSync(agPath);
}

fs.symlinkSync('ag-' + os + '-' + arch, agPath);
