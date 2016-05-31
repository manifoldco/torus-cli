'use strict';

var envsInfo = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var client = require('../api/client').create();
var Session = require('../session');
var validate = require('../validate');

envsInfo.output = {};

/**
 * Success
 *
 * @param {object} payload - Response object
 */
envsInfo.output.success = output.create(function (payload) {
  var env = payload.body;

  console.log('Environment ' + env.body.name + ':');
  console.log('');
  console.log('Nothing here yet! Work in progress :)');
});

envsInfo.output.failure = output.create(function () {
  console.log('Retrieval of envs failed!');
});

/**
 * List envs
 *
 * @param {object} ctx - Command context
 */
envsInfo.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.session instanceof Session)) {
      throw new TypeError('Session object missing on Context');
    }

    client.auth(ctx.session.token);

    var envName = ctx.params[0];

    if (!envName) {
      return reject(new Error('Environment [name] is required.'));
    }

    var envNameErr = validate.slug(envName);

    if (!_.isBoolean(envNameErr)) {
      return reject(envNameErr);
    }

    var options = ctx.options;
    var serviceName = options && options.service && options.service.value;

    if (!serviceName) {
      return reject(new Error('Service [-s --service]'));
    }

    var serviceNameErr = validate.slug(serviceName);

    if (!_.isBoolean(serviceNameErr)) {
      return reject(serviceNameErr);
    }

    return client.get({ url: '/services/' + serviceName })
      .then(function (servicePayload) {
        var service = servicePayload.body;

        return client.get({ url: '/envs', qs: { owner_id: service.id } });
      })
      .catch(reject);
  });
};
