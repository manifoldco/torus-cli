'use strict';

var envCreate = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var Session = require('../session');

var client = require('../api/client').create();

envCreate.output = {};

envCreate.output.success = output.create(function () {
  console.log('Environment created');
});

envCreate.output.failure = output.create(function () {
  console.log('Environment creation failed, please try again');
});

envCreate.questions = function (serviceNames) {
  return [
    [
      {
        name: 'name',
        message: 'Environment name',
        validate: validate.slug
      },
      {
        type: 'list',
        name: 'service',
        message: 'Service',
        choices: serviceNames
      }
    ]
  ];
};

var validator = validate.build({
  name: validate.slug,
  service: validate.slug
});

/**
 * Create prompt for env create
 *
 * @param {object} ctx - Command context
 */
envCreate.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var params = ctx.params || [];

    var data = {
      name: params[0],
      service: ctx.options.service.value
    };

    var orgName = ctx.options.org.value;

    if (!orgName) {
      return reject(new Error('--org is (temporarily) required.'));
    }

    var validOrgName = validate.slug(orgName);

    if (!_.isBoolean(validOrgName)) {
      return reject(new Error(validOrgName));
    }

    client.auth(ctx.session.token);

    var serviceData;
    var org;
    return client.get({
      url: '/orgs',
      qs: { name: orgName }
    }).then(function (res) {
      org = res.body && res.body[0];

      if (!org) {
        throw new Error('Env not created; invalid org provided');
      }

      return client.get({
        url: '/services',
        qs: { org_id: org.id }
      }).then(function (results) {
        serviceData = results.body;
        return _.map(serviceData, 'body.name');

      // Prompt for values, filling in defaults for supplied values
      }).then(function (serviceNames) {
        if (!serviceNames.length) {
          throw new Error('Must create service before env');
        }

        // If sufficient data supplied, return early
        if (data.name && data.service) {
          data = _.omitBy(data, _.isUndefined);
          data = _.mapValues(data, _.toString);

          // Validate inputs from params/options
          var errors = validator(data);
          if (errors.length) {
            return reject(errors[0]);
          }

          return data;
        }

        // Otherwise prompt for missing values
        return envCreate._prompt(data, serviceNames).then(function (userInput) {
          userInput = _.omitBy(_.extend({}, data, userInput), _.isUndefined);
          return userInput;
        });

      // Map the item selected to its ID
      })
      .then(function (userInput) {
        var service = _.find(serviceData, function (s) {
          return s.body.name === userInput.service;
        });

        if (!service) {
          throw new Error('Unknown service: ' + userInput.service);
        }

        return {
          body: {
            name: userInput.name,
            project_id: service.body.project_id,
            org_id: org.id
          }
        };
      })
      .then(function (envData) {
        return envCreate._execute(ctx.session, envData);
      });
    })
    .then(resolve)
    .catch(reject);
  });
};

/**
 * Create prompt promise
 */
envCreate._prompt = function (defaults, serviceNames) {
  var prompt = new Prompt({
    stages: envCreate.questions,
    defaults: defaults,
    questionArgs: [
      serviceNames
    ]
  });

  return prompt.start();
};

/**
 * Run the create request
 *
 * @param {Session} session
 * @param {object} data
 */
envCreate._execute = function (session, data) {
  return new Promise(function (resolve, reject) {
    if (!(session instanceof Session)) {
      throw new TypeError('Session object missing on Context');
    }

    client.post({
      url: '/envs',
      json: data
    })
    .then(resolve)
    .catch(reject);
  });
};
