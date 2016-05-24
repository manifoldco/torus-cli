'use strict';

var serviceCreate = exports;

var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

serviceCreate.output = {};

/**
 * Success output
 */
serviceCreate.output.success = output.create(function() {
  console.log('Service created.');
});

/**
 * Failure output
 */
serviceCreate.output.failure = output.create(function() {
  console.log('Service creation failed, please try again');
});

/**
 * Service prompt questions
 */
serviceCreate.questions = function() {
  return [
    [
      {
        name: 'name',
        message: 'Service name',
        validate: validate.slug,
      }
    ]
  ];
};

var validator = validate.build({
  name: validate.slug
});

/**
 * Create prompt for service create
 *
 * @param {object} ctx - Command context
 */
serviceCreate.execute = function(ctx) {
  return new Promise(function(resolve, reject) {
    var retrieveInput;
    var name = ctx.params[0];

    var errors = validator({ name: name });
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    if (name) {
      retrieveInput = Promise.resolve({
        name: name
      });
    } else {
      retrieveInput = serviceCreate._prompt();
    }

    retrieveInput.then(function(userInput) {
      return serviceCreate._execute(ctx.token, userInput);
    }).then(resolve).catch(reject);
  });
};

/**
 * Attempt to create service with supplied input
 *
 * @param {string} token - Auth token
 * @param {object} userInput
 */
serviceCreate._execute = function(token, userInput) {
  return new Promise(function(resolve, reject) {
    client.auth(token);

    if (!client.authToken) {
      return reject(new Error('must authenticate first'));
    }

    client.post({
      url: '/services',
      json: {
        body: {
          name: userInput.name
        }
      }
    }).then(resolve).catch(reject);
  });
};

/**
 * Create prompt promise
 */
serviceCreate._prompt = function() {
  var prompt = new Prompt({
    stages: serviceCreate.questions
  });

  return prompt.start();
};
