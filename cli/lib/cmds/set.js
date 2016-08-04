'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
var setCred = require('../credentials/set');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var set = new Command(
  'set <name> <value>',
  'set the name for the given service and environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      setCred.execute(ctx).then(function (cred) {
        setCred.output.success(ctx, cred);
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
set.hook('pre', target());

flags.add(set, 'org', {
  description: 'the org the secret will belong to'
});
flags.add(set, 'project', {
  description: 'the project the secret will belong to'
});
flags.add(set, 'environment', {
  description: 'the environment the secret will belong to'
});
flags.add(set, 'service', {
  description: 'the service the secret will belong to'
});
flags.add(set, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

module.exports = set;
