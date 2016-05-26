'use strict';

var envCreate = exports;

var Promise = require('es6-promise').Promise;
var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var services = require('../services/list');
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
    var options = ctx.options || {};
    var params = ctx.params || [];
    var service = options.service && options.service.value;

    var data = {
      name: params[0],
      service: service
    };

    var serviceData;
    services.execute(ctx).then(function (results) {
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
      var serviceId = _.map(_.filter(serviceData, function (s) {
        return s.body.name === userInput.service;
      }), 'id');

      // Swap for owner_id since we know the exact one
      userInput.owner_id = serviceId[0];
      delete userInput.service;
      return userInput;

    // Create the env in the
    })
    .then(function (userInput) {
      return envCreate._execute(ctx.session, userInput);
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

    client.auth(session.token);

    client.post({
      url: '/envs',
      json: {
        body: data
      }
    })
    .then(resolve)
    .catch(reject);
  });
};
