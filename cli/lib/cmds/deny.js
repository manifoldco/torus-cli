'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var deny = require('../policies/deny');
var auth = require('../middleware/auth');

var command = new Command(
  'deny [crudl] [path] [team]',
  'Explicitly deny a team permissions access to a secret[\'s]',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      deny.execute(ctx).then(function (payload) {
        deny.output.success(null, payload);

        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        deny.output.failure();
        reject(err);
      });
    });
  }
);

command.hook('pre', auth());

module.exports = command;
