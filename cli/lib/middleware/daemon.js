'use strict';

var daemon = require('../daemon');

module.exports.preHook = function () {
  return function (ctx) {
    return new Promise(function(resolve, reject) {
      if (!ctx.config) {
        return reject(new Error('Must have config property on Context'));
      }

      daemon.get(ctx.config).then(function(d) {
        if (d) {
          ctx.daemon = d;
          return resolve();
        }

        return daemon.start(ctx.config).then(function(d) {
          ctx.daemon = d;
          resolve();
        });
      }).catch(reject);
    });
  };
};

module.exports.postHook = function () {
  return function (ctx) {
    if (!ctx.config || !ctx.daemon) {
      return Promise.reject(
        new Error('Must have config and daemon on Content'));
    }

    return ctx.daemon.disconnect();
  };
};
