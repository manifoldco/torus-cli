#!/usr/bin/env node
var path = require('path');

// os and arch restrictions are handled by the package.json
var os = process.platform;
var arch = 'amd64';

// Select the right binary for this platform, then exec it with the original
// arguments. This is a true exec(3), which will take over the pid, env, and
// file descriptors.
var torusPath = path.join(__dirname, '../bin/torus-' + os + '-' + arch);
if (os === 'win32') {
  torusPath = path.join(__dirname, '../bin/torus.exe');
}

try {
  var kexec = require('kexec');
  kexec(torusPath, process.argv.slice(2));
} catch (err) {
  if (err.code !== 'MODULE_NOT_FOUND') {
    console.error('Could not leverage kexec due to error: ' + err.message);
  }

  var spawn = require('child_process').spawn;
  var proc = spawn(torusPath, process.argv.slice(2), { stdio: 'inherit' });
  proc.on('exit', function (code, signal) {
    process.on('exit', function () {
      if (signal) {
        process.kill(process.pid, signal);
      } else {
        process.exit(code);
      }
    });
  });
}
