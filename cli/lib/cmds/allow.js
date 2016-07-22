'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var allow = require('../access/allow');
var auth = require('../middleware/auth');

var command = new Command(
  'allow',
  'Grant a team permissions on a resource',
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
