'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
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

flags.add(set, 'org', {
  description: 'the org the credential will belong too'
});
flags.add(set, 'service', {
  description: 'the service the credential will belong too'
});
flags.add(set, 'environment', {
  description: 'the environment the credential will belong too'
});
flags.add(set, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

module.exports = set;
