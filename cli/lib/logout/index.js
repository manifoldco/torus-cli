'use strict';

var Promise = require('es6-promise').Promise;

var Session = require('../session');
var client = require('../api/client').create();

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

    client.auth(ctx.session.token);

    function resetClient() {
      client.reset();
      ctx.session = null;
    }

    return Promise.all([
      ctx.daemon.logout(),
      client.delete({ url: '/tokens/' + ctx.session.token })
    ]).then(function () {
      resetClient();
      resolve();
    }).catch(function (err) {
      resetClient();
      reject(err);
    });
  });
};
