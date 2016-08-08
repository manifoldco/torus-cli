'use strict';

var Promise = require('es6-promise').Promise;
var Command = require('../cli/command');

var flags = require('../flags');
var viewCred = require('../credentials/view');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var view = new Command(
  'view',
  'View secrets for the current service and environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      viewCred.execute(ctx).then(function (creds) {
        viewCred.output.success(ctx, creds);
        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        viewCred.output.failure(err);
        reject(err);
      });
    });
  }
);

view.hook('pre', auth());
view.hook('pre', target());

flags.add(view, 'org', {
  description: 'Org to which the secrets belongs'
});
flags.add(view, 'project', {
  description: 'Project to which the secrets belongs'
});
flags.add(view, 'service', {
  description: 'Service to which the secret belongs'
});
flags.add(view, 'environment', {
  description: 'Environment to which the secret belongs'
});
flags.add(view, 'instance', {
  description: 'Instance of the service belonging to the current user'
});

view.option('-v, --verbose', 'list the sources of the values');

module.exports = view;
