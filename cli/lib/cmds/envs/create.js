'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var flags = require('../../flags');
var envs = require('../../envs');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'envs:create [name]',
  'create an environment for a service',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      envs.create.execute(ctx).then(function () {
        envs.create.output.success();
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        envs.create.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');
flags.add(cmd, 'project', {
  description: 'the project this environment will belong to'
});

module.exports = cmd;
