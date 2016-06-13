'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var envs = require('../../envs');
var auth = require('../../middleware/auth');

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

cmd.option(
  '-s, --service [service]',
  'service the environment will belong too',
  undefined
);

cmd.option(
  '-o, --org [org]',
  'Specify an organization',
  null
);

cmd.hook('pre', auth());

module.exports = cmd;
