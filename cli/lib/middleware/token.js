'use strict';

var Promise = require('es6-promise').Promise;

module.exports.preHook = function() {
  return function(ctx) {
    if (process.env.NO_DAEMON) {
      return Promise.resolve();
    }

    return ctx.daemon.get('token').then(function(result) {
      ctx.token = result.token;
    });
  };
};
