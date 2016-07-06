'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
var status = require('../context/status');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var cmd = new Command(
  'status',
  'shows your current arigato status based on your identity and CWD',
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
  description: 'the service to add to context'
});

module.exports = cmd;
