'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
var unsetCred = require('../credentials/unset');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var unset = new Command(
  'unset <name>',
  'Remove a secret from a service and environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      unsetCred.execute(ctx).then(function (cred) {
        unsetCred.output.success(ctx, cred);
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
unset.hook('pre', target());

flags.add(unset, 'org', {
  description: 'Organization from which the secret will be removed'
});
flags.add(unset, 'project', {
  description: 'Project from which the secret will be removed'
});
flags.add(unset, 'environment', {
  description: 'Environment from which the secret will be removed'
});
flags.add(unset, 'service', {
  description: 'Service from which the secret will be removed'
});
flags.add(unset, 'user', {
  description: 'User who can access the secret within the environment'
});
flags.add(unset, 'instance', {
  description: 'Instance of the service belonging to the current user'
});

module.exports = unset;
