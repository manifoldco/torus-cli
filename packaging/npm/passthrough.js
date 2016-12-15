#!/usr/bin/env node
var path = require('path');
var kexec = require('kexec');

// os and arch restrictions are handled by the package.json
var os = process.platform;
var arch = 'amd64';

// Select the right binary for this platform, then exec it with the original
// arguments. This is a true exec(3), which will take over the pid, env, and
// file descriptors.
var torusPath = path.join(__dirname, '../bin/torus-' + os + '-' + arch);
kexec(torusPath, process.argv.slice(2));
