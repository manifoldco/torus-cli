'use strict';

var path = require('path');

function Config(arigatoRoot, version) {
  var socketPath = path.join(arigatoRoot, 'daemon.socket');

  this.arigatoRoot = arigatoRoot;
  this.socketUrl = 'http://unix:' + socketPath + ':';
  this.registryUrl = 'https://api.arigato.sh';
  this.pidPath = path.join(arigatoRoot, 'daemon.pid');
  this.rcPath = path.join(process.env.HOME, '.arigatorc');
  this.version = version;
}

module.exports = Config;
