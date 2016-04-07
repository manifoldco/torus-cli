'use strict';

var Command = require('../../cli/command');

module.exports = new Command(
  'envs',
  'list all environments for this application',
  function () {
    console.log('listing all envs here');
  }
);
