'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
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

flags.add(unset, 'org', {
  description: 'the org the credential will belong to'
});
flags.add(unset, 'project', {
  description: 'the project the credential will belong to'
});
flags.add(unset, 'environment', {
  description: 'the environment the credential will belong to'
});
flags.add(unset, 'service', {
  description: 'the service the credential will belong to'
});
flags.add(unset, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

module.exports = unset;
