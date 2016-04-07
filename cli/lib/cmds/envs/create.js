'use strict';

var Command = require('../../cli/command');

module.exports = new Command(
  'envs:create [name]',
  'create an environment for the current application',
  function (ctx) {
    console.log('creating an environment!', ctx.params);
  }
);
