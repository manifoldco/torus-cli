'use strict';

var Promise = require('es6-promise').Promise;

var Session = require('../session');

/**
 * Attempt logout
 *
 * @param {object} ctx - Prompt context
 */
module.exports = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.session instanceof Session)) {
      throw new TypeError('Session object not on Context object');
    }

    function resetClient() {
      ctx.api.reset();
      ctx.session = null;
    }

    return Promise.all([
      ctx.daemon.logout(),
      ctx.api.tokens.remove({}, { auth_token: ctx.session.token })
    ]).then(function () {
      resetClient();
      resolve();
    }).catch(function (err) {
      resetClient();
      reject(err);
    });
  });
};
