'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var allow = require('../policies/allow');
var auth = require('../middleware/auth');

var example = 'crudl /knotty-buoy/landing-page/production/*/*/*/* api-ops\n\n  ';
example += 'Gives the api-ops team create, read, update, delete, and list ';
example += 'access for the project landing-page in the production environment.';

var command = new Command(
  'allow <crudl> <path> <team>',
  'Grant a team permission to access specific resources',
  example,
  function (ctx) {
    return new Promise(function (resolve, reject) {
      allow.execute(ctx).then(function (payload) {
        allow.output.success(null, payload);
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        allow.output.failure();
        reject(err);
      });
    });
  }
);

command.hook('pre', auth());

module.exports = command;
