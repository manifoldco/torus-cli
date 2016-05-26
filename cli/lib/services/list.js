'use strict';

var servicesList = exports;
var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var Session = require('../session');
var output = require('../cli/output');
var client = require('../api/client').create();

servicesList.output = {};

servicesList.output.success = output.create(function (services) {
  var length = services.length;

  var msg = services.map(function (service) {
    return _.padStart(service.body.name, service.body.name.length + 1);
  }).join('\n');

  console.log(' total (' + length + ')\n ---------\n' + msg);
});

servicesList.output.failure = output.create(function () {
  console.log('Retrieval of services failed!');
});

/**
 * List services
 *
 * @param {object} ctx
 */
servicesList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.session instanceof Session)) {
      throw new TypeError('Session object not on Context');
    }

    client.auth(ctx.session.token);
    return client.get({ url: '/services' })
      .then(resolve)
      .catch(reject);
  });
};
