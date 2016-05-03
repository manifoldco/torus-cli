'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var login = require('../../login');

module.exports = new Command(
  'login',
  'log in to your Arigato account',
  function(ctx) {
    return new Promise(function(resolve, reject) {
      login.execute(ctx).then(function() {
        login.output.success();
        resolve();

      // Account creation failed
      }).catch(function(err) {
        err.type = err.type || 'unknown';
        login.output.failure();
        reject(err);
      });
    });
  }
);
