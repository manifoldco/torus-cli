'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var envs = require('../../envs');
var auth = require('../../middleware/auth');

var listEnv = new Command(
  'envs',
  'list environments',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      envs.list.execute(ctx).then(function (payload) {
        envs.list.output.success(null, payload);

        resolve(true);
      }).catch(function (err) {
        envs.list.output.failure();
        reject(err);
      });
    });
  }
);

listEnv.option(
  '-s, --service [service]',
  'List environments for a particular service',
  null
);

listEnv.hook('pre', auth());

module.exports = listEnv;
