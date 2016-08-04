'use strict';

var Promise = require('es6-promise').Promise;
var Command = require('../cli/command');

var flags = require('../flags');
var viewCred = require('../credentials/view');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var view = new Command(
  'view',
  'view secrets for the current service and environment',
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
  description: 'the org the secrets belongs to'
});
flags.add(view, 'project', {
  description: 'the project the secrets belongs to'
});
flags.add(view, 'service', {
  description: 'the service the secrets belong to'
});
flags.add(view, 'environment', {
  description: 'the environment the secrets belong to'
});
flags.add(view, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

view.option('-v, --verbose', 'list the sources of the values');

module.exports = view;
