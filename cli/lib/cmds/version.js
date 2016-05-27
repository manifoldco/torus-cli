'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var version = new Command(
  'version',
  'displays versions of the cli and daemon',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      ctx.daemon.version().then(function (msg) {
        console.log('CLI Version: ' + ctx.config.version);
        console.log('Daemon Version: ' + msg.version);

        resolve();
      }).catch(reject);
    });
  }
);

module.exports = version;
