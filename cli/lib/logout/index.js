'use strict';

var Promise = require('es6-promise').Promise;
var client = require('../api/client').create();

/**
 * Attempt logout
 *
 * @param {object} ctx - Prompt context
 */
module.exports = function(ctx) {
  client.auth(ctx.token);

  var resetClient = client.reset.bind(client);

  return Promise.all([
    ctx.daemon.logout(),
    client.delete({ url: '/session/' + ctx.token })
  ]).then(resetClient)
    .catch(function(err) {
      resetClient();
      throw err;
    });
};
