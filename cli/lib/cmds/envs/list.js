'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var envs = require('../../envs');
var auth = require('../../middleware/auth');

var cmd = new Command(
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

cmd.option(
  '-s, --service [service]',
  'List environments for a particular service',
  null
);

cmd.option(
  '-o, --org [org]',
  'Specify an organization',
  null
);

cmd.hook('pre', auth());

module.exports = cmd;
