'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
var setCred = require('../credentials/set');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var example = 'port 3001 -o nick-admin -p landing-page -s www -e dev\n\n  ';
example += 'Sets the secret \'port\' to 3001 in a service in an organization\'s project';

var set = new Command(
  'set <name> <value>',
  'Set a secret for a service and environment',
  example,
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
  description: 'Organization to which the secret will belong'
});
flags.add(set, 'project', {
  description: 'Project to which the secret will belong'
});
flags.add(set, 'environment', {
  description: 'Environment to which the secret will belong'
});
flags.add(set, 'service', {
  description: 'Service to which the secret will belong'
});
flags.add(set, 'instance', {
  description: 'Instance of the service belonging to the current user'
});

module.exports = set;
