#!/usr/bin/env node
var fs = require('fs');

if (['linux', 'darwin'].indexOf(process.platform) === -1) {
  console.log('ag only supports linux and darwin amd64');
  throw new Error('Incompatible Platform: ' + process.platform);
}

var os = process.platform;
var arch = 'amd64';

fs.unlinkSync(__dirname + '/../bin/ag');
fs.symlinkSync('ag-' + os + '-' + arch, __dirname + '/../bin/ag');
