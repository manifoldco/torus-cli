'use strict';

var Promise = require('es6-promise').Promise;
var daemon = require('../daemon');

module.exports.preHook = function () {
  return function (ctx) {
    if (process.env.NO_DAEMON) {
      return Promise.resolve();
    }

    return new Promise(function (resolve, reject) {
      if (!ctx.config) {
        return reject(new Error('Must have config property on Context'));
      }

      return daemon.get(ctx.config).then(function (d) {
        if (d) {
          ctx.daemon = d;
          return resolve();
        }

        return daemon.start(ctx.config).then(function (dd) {
          ctx.daemon = dd;
          return resolve();
        });
      }).catch(reject);
    });
  };
};

module.exports.postHook = function () {
  return function (ctx) {
    if (process.env.NO_DAEMON) {
      return Promise.resolve();
    }

    if (!ctx.config || !ctx.daemon) {
      return Promise.reject(
        new Error('Must have config and daemon on Content'));
    }

    return ctx.daemon.disconnect();
  };
};
