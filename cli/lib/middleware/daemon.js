'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var daemon = require('../daemon');
var api = require('../api');

module.exports.preHook = function () {
  return function (ctx) {
    var config = ctx.config;

    ctx.api = api.build({
      registryUrl: _.get(ctx, 'ctx.prefs.values.core.registry_uri', config.registryUrl),
      socketUrl: config.socketUrl
    });

    return new Promise(function (resolve, reject) {
      if (!config) {
        return reject(new Error('Must have config property on Context'));
      }

      // The daemon won't start in some circumstances without a public_key_file
      // preference set. We can't set a preference if we require the daemon to
      // be running!
      if (ctx.cmd.group === 'prefs') {
        return resolve();
      }

      return daemon.status(config)
        .then(function (status) {
          // If the daemon isn't running, start it
          if (!status.exists) return daemon.start(config);
          return true;
        })
        .then(function () {
          return ctx.api.versionApi.get();
        })
        .then(function (res) {
          if (config.version === res.daemon.version) return res;

          console.log('The ag-daemon version is out of date.');
          console.log();
          console.log(
            'The daemon is being restarted, you will need to login again\n');

          // If the daemon isn't running the same version, restart it
          return daemon.restart(config).then(function () {
            return ctx.api.versionApi.get();
          });
        })
        .then(function (res) {
          if (config.version !== res.daemon.version) {
            throw new Error('Wrong version of daemon running, check for zombie ag-daemon process');
          }

          return true;
        })
        .then(resolve)
        .catch(reject);
    });
  };
};
