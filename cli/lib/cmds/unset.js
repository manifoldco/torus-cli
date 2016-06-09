'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var unsetCred = require('../credentials/unset');
var auth = require('../middleware/auth');

var unset = new Command(
  'unset <name>',
  'unset the name for the given service and environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      unsetCred.execute(ctx).then(function (cred) {
        unsetCred.output.success(cred);
        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        unsetCred.output.failure(err);
        reject(err);
      });
    });
  }
);

unset.hook('pre', auth());

unset.option(
  '-s, --service [service]',
  'the service the credential belongs too',
  undefined
);

unset.option(
  '-e, --environment [environment]',
  'the environment the credential belongs too',
  undefined
);

unset.option(
  '-i, --instance [name]',
  'the instance of the service belonging to the current user',
  '1'
);

module.exports = unset;
