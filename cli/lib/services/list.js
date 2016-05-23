'use strict';

var servicesList = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var client = require('../api/client').create();

servicesList.output = {};

servicesList.output.success = output.create(function(services) {
  var length = services.length;

  var msg = services.map(function(service) {
    return _.padStart(service.body.name, 4);
  }).join('\n');

  console.log(' total (' + length + ')\n ---------\n' + msg);
});

servicesList.output.failure = output.create(function() {
  console.log('Retrieval of services failed!');
});

/**
 * List services
 *
 * @param {string} token - Auth token
 * @param {object} userInput
 */
servicesList.execute = function(ctx) {
  return new Promise(function(resolve, reject) {
    if (!ctx.token) {
      return reject(new Error('must authenticate first'));
    }

    client.auth(ctx.token);
    return client.get({ url: '/services' })
      .then(resolve)
      .catch(reject);
  });
};
