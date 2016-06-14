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

  var envsByOwner = _.groupBy(envs, 'body.project_id');

  _.each(services, function (service) {
    var serviceEnvs = envsByOwner[service.body.project_id] || [];
    var serviceName = service.body.name;

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

    var orgName = ctx.options.org.value;

    if (!orgName) {
      return reject(new Error('--org is (temporarily) required.'));
    }

    var validOrgName = validate.slug(orgName);

    if (!_.isBoolean(validOrgName)) {
      return reject(new Error(validOrgName));
    }

    var serviceName = ctx.options.service.value;

    if (!serviceName) {
      return reject(new Error('--service is (temporarily) required.'));
    }

    var validServiceName = validate.slug(serviceName);

    if (!_.isBoolean(validServiceName)) {
      return reject(validServiceName);
    }

    client.auth(ctx.session.token);

    return client.get({
      url: '/orgs',
      qs: { name: orgName }
    }).then(function (res) {
      var org = res.body && res.body[0];

      if (!_.isObject(org)) {
        return reject(new Error('The org could not be found'));
      }

      return client.get({
        url: '/services',
        qs: {
          org_id: org.id,
          name: serviceName
        }
      }).then(function (servicePayload) {
        // TODO: No services? Prompt them to create one first.
        var services = servicePayload.body;
        if (!Array.isArray(services)) {
          return reject(new Error('API returned invalid services list'));
        }

        return client.get({
          url: '/envs',
          qs: {
            org_id: org.id,
            project_id: services[0].body.project_id
          }
        }).then(function (envsPayload) {
          var envs = envsPayload.body;
          if (!Array.isArray(envs)) {
            return reject(new Error('API returned invalid envs list'));
          }

          return {
            services: services,
            envs: envs
          };
        });
      });
    }).then(resolve)
      .catch(reject);
  });
};
