'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

module.exports = new Command(
  'services',
  'list services for your org',
  function(/*ctx*/) {
    return new Promise(function(resolve) {
      resolve();
    });
  }
);
