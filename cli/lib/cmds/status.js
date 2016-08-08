'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
var status = require('../context/status');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var cmd = new Command(
  'status',
  'Show the current Arigato status associated with your account and project',
  '-s www -e development\n\n  Shows the status of the \'www\' service in a development environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      status.execute(ctx).then(function (tgt) {
        status.output.success(null, ctx, tgt);
        resolve();
      }).catch(function (err) {
        status.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'service', {
  description: 'Service to add to context'
});

flags.add(cmd, 'environment', {
  description: 'Environment to add to context'
});

module.exports = cmd;
