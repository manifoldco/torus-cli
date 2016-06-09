'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var setCred = require('../credentials/set');
var auth = require('../middleware/auth');

var set = new Command(
  'set <name> <value>',
  'set the name for the given service and environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      setCred.execute(ctx).then(function (cred) {
        setCred.output.success(cred);
        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        setCred.output.failure(err);
        reject(err);
      });
    });
  }
);

set.hook('pre', auth());

set.option(
  '-s, --service [service]',
  'the service the credential will belong too',
  undefined
);

set.option(
  '-e, --environment [environment]',
  'the environment the credential will belong too',
  undefined
);

set.option(
  '-i, --instance [name]',
  'the instance of the service belonging to the current user',
  '1'
);

module.exports = set;
