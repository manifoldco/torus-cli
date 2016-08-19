'use strict';

var Command = require('../cli/command');

var version = new Command(
  'version',
  'Display a list of versions of the CLI and daemon',
  function (ctx) {
    return ctx.api.versionApi.get().then(function (msg) {
      console.log();
      console.log(' Versions');
      console.log(' ---------');
      console.log(' CLI:      ' + ctx.config.version);
      console.log(' Daemon:   ' + msg.daemon.version);
      console.log(' Registry: ' + msg.registry.version);
      console.log();
    });
  }
);

module.exports = version;
