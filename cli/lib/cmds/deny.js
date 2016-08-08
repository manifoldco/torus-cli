'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var deny = require('../policies/deny');
var auth = require('../middleware/auth');

var example = 'crudl /knotty-buoy/landing-page/production/*/*/*/* api-ops\n\n  ';
example += 'Removes api-ops team access to create, read, update, delete, and list ';
example += 'for the project landing-page in the production environment.';

var command = new Command(
  'deny <crudl> <path> <team>',
  'Deny a team permission to access specific resources',
  example,
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
