'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var version = new Command(
  'version',
  'Display a list of versions of the CLI and daemon',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      ctx.api.versionApi.get().then(function (msg) {
        console.log('CLI Version: ' + ctx.config.version);
        console.log('Daemon Version: ' + msg.version);

        resolve();
      }).catch(reject);
    });
  }
);

module.exports = version;
