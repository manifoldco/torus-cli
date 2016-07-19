'use strict';

var Promise = require('es6-promise').Promise;
var daemon = require('../daemon');

module.exports.preHook = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      if (!ctx.config) {
        return reject(new Error('Must have config property on Context'));
      }
      return daemon.status(ctx.config).then(function (status) {
        if (status.exists) {
          return resolve();
        }

        return daemon.start(ctx.config).then(function () {
          return resolve();
        });
      }).catch(reject);
    });
  };
};
