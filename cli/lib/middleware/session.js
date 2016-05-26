'use strict';

var Session = require('../session');

module.exports = function () {
  return function (ctx) {
    if (process.env.NO_DAEMON) {
      return true;
    }

    return ctx.daemon.get().then(function (result) {
      ctx.session = (!result || !result.token || !result.passphrase) ?
       null : new Session(result);

      return true;
    });
  };
};
