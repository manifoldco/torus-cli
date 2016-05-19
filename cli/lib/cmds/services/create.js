'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var services = require('../../services');

module.exports = new Command(
  'services:create [name]',
  'create a new service for your org',
  function(ctx) {
    return new Promise(function(resolve, reject) {
      services.create.execute(ctx).then(function() {
        services.create.output.success();
        resolve();

      // Service creation failed
      }).catch(function(err) {
        err.type = err.type || 'unknown';
        services.create.output.failure(err);
        reject(err);
      });
    });
  }
);
