'use strict';

var path = require('path');

function Config(arigatoRoot, version) {
  this.arigatoRoot = arigatoRoot;
  this.socketPath = path.join(arigatoRoot, 'daemon.socket');
  this.pidPath = path.join(arigatoRoot, 'daemon.pid');
  this.version = version;
}

module.exports = Config;
