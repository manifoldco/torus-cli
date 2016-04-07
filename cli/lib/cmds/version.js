'use strict';

var Command = require('../cli/command');

var version = new Command(
  'version',
  'displays versions of the cli and daemon',
  function (ctx) {
    console.log('Version: '+ctx.program.version);
    return Promise.resolve(true);
  }
);

module.exports = version;
