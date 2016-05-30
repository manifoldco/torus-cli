'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var envs = require('../../envs');
var auth = require('../../middleware/auth');

var createEnv = new Command(
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

createEnv.option(
  '-s, --service [service]',
  'service the environment will belong too ',
  undefined
);

createEnv.hook('pre', auth());

module.exports = createEnv;
