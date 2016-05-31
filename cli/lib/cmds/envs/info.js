'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var envs = require('../../envs');
var auth = require('../../middleware/auth');

var infoEnv = new Command(
  'envs:info [name]',
  'view information about an environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      envs.info.execute(ctx).then(function (service) {
        envs.info.output.success(null, service);

        resolve(true);
      }).catch(function (err) {
        envs.info.output.failure();
        reject(err);
      });
    });
  }
);

infoEnv.option(
  '-s, --service [service]',
  '[Required] Specify the service associated with the environment',
  null
);

infoEnv.hook('pre', auth());

module.exports = infoEnv;
