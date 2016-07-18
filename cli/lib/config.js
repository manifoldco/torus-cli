'use strict';

var path = require('path');

function Config(arigatoRoot, version) {
  var socketPath = path.join(arigatoRoot, 'daemon.socket');

  this.arigatoRoot = arigatoRoot;
  this.socketUrl = 'http://unix:' + socketPath + ':';
  this.pidPath = path.join(arigatoRoot, 'daemon.pid');
  this.version = version;
}

module.exports = Config;
