'use strict';

var envCreate = exports;

var _ = require('lodash');

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var Promise = require('es6-promise').Promise;
var services = require('../services/list');
var client = require('../api/client').create();

envCreate.output = {};

envCreate.output.success = output.create(function() {
  console.log('Environment created');
});

envCreate.output.failure = output.create(function() {
  console.log('Environment creation failed, please try again');
});

envCreate.questions = function(services) {
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
        choices: services
      }
    ]
  ];
};

var validator = {
  id: validate.build({
    name: validate.slug,
  }),
  name: validate.build({
    name: validate.slug,
    service: validate.slug,
  }),
};

/**
 * Create prompt for env create
 *
 * @param {object} ctx - Command context
 */
envCreate.execute = function(ctx) {
  return new Promise(function(resolve, reject) {

    var options = ctx.options || {};
    var params = ctx.params || [];
    var service = options.service && options.service.value;

    var data = {
      name: params[0],
      service: service,
    };

    var serviceData;
    services.execute(ctx).then(function(services) {
      serviceData = services.body;
      return _.map(serviceData, 'body.name');

    // Prompt for values, filling in defaults for supplied values
    }).then(function(serviceNames) {
      if (!serviceNames.length) {
        throw new Error('Must create service before env');
      }

      // If sufficient data supplied, return early
      if (data.name && data.service) {
        data = _.omitBy(data, _.isUndefined);
        data = _.mapValues(data, _.toString);

        // Validate inputs from params/options
        var validateData = data.service? validator.name : validator.id;
        var errors = validateData(data);
        if (errors.length) {
          return reject(errors[0]);
        }

        return data;
      }

      // Otherwise prompt for missing values
      return envCreate._prompt(data, serviceNames).then(function(userInput) {
        userInput = _.omitBy(_.extend({}, data, userInput), _.isUndefined);
        return userInput;
      });

    // Map the item selected to its ID
    }).then(function(userInput) {
      var serviceId = _.map(_.filter(serviceData, function(service) {
        return service.body.name === userInput.service;
      }), 'id');

      // Swap for owner_id since we know the exact one
      userInput.owner_id = serviceId[0];
      delete userInput.service;
      return userInput;

    // Create the env in the
    }).then(function(userInput) {

      return envCreate._execute(ctx.token, userInput);

    }).then(resolve).catch(reject);
  });
};

/**
 * Create prompt promise
 */
envCreate._prompt = function(defaults, services) {
  var prompt = new Prompt({
    stages: envCreate.questions,
    defaults: defaults,
    questionArgs: [
      services
    ]
  });

  return prompt.start();
};

/**
 * Run the create request
 *
 * @param {string} token
 * @param {object} data
 */
envCreate._execute = function(token, data) {
  return new Promise(function(resolve, reject) {

    // Authenticate the API client
    client.auth(token);

    if (!client.authToken) {
      throw new Error('must authenticate first');
    }

    client.post({
      url: '/envs',
      json: {
        body: data
      }
    }).then(resolve).catch(reject);
  });
};
