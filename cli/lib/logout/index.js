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

    return ctx.api.logout.post()
    .then(function () {
      resetClient();
      resolve();
    })
    .catch(function (err) {
      resetClient();
      reject(err);
    });
  });
};
