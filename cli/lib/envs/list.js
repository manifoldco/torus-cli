'use strict';

var envsList = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var validate = require('../validate');
var output = require('../cli/output');
var client = require('../api/client').create();
var Session = require('../session');

envsList.output = {};

/**
 * Success
 *
 * @param {object} payload - Response object
 */
envsList.output.success = output.create(function (payload) {
  var envs = payload.envs;
  var services = payload.services;

  var envsByOwner = _.groupBy(envs, 'body.owner_id');

  _.each(services, function (service) {
    var serviceEnvs = envsByOwner[service.id];
    var serviceName = service.body.name;

    if (!serviceEnvs) return;

    var msg = ' ' + serviceName + ' service (' + serviceEnvs.length + ')\n';

    msg += ' ' + new Array(msg.length - 1).join('-') + '\n'; // underline
    msg += serviceEnvs.map(function (env) {
      return _.padStart(env.body.name, env.body.name.length + 1);
    }).join('\n');

    if (envs !== _.findLast(serviceEnvs)) {
      msg += '\n';
    }

    console.log(msg);
  });
});

envsList.output.failure = output.create(function () {
  console.log('Retrieval of envs failed!');
});

/**
 * List envs
 *
 * @param {object} ctx - Command context
 */
envsList.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.session instanceof Session)) {
      throw new TypeError('Session object missing on Context');
    }

    client.auth(ctx.session.token);

    var options = ctx.options;
    var serviceName = options && options.service && options.service.value;
    var serviceOpts = { url: '/services' };

    // Lookup by service name if one is provided and valid
    if (serviceName) {
      var err = validate.slug(serviceName);

      if (!_.isBoolean(err)) {
        return reject(err);
      }

      serviceOpts.url += '/' + serviceName;
    }

    return client.get(serviceOpts)
      .then(function (servicePayload) {
        // TODO: No services? Prompt them to create one first.

        // Get environments associated with each service
        var findEnvs = servicePayload.body.map(function (service) {
          return client.get({ url: '/envs', qs: { owner_id: service.id } });
        });

        return Promise.all(findEnvs).then(function (envsPayload) {
          return resolve({
            services: servicePayload.body,
            envs: _.flatMap(envsPayload, function (envPayload) {
              return envPayload.body;
            })
          });
        });
      })
      .catch(reject);
  });
};
