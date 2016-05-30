'use strict';

var serviceCreate = exports;

var Promise = require('es6-promise').Promise;

var Session = require('../session');
var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

serviceCreate.output = {};

/**
 * Success output
 */
serviceCreate.output.success = output.create(function () {
  console.log('Service created.');
});

/**
 * Failure output
 */
serviceCreate.output.failure = output.create(function () {
  console.log('Service creation failed, please try again');
});

/**
 * Service prompt questions
 */
serviceCreate.questions = function () {
  return [
    [
      {
        name: 'name',
        message: 'Service name',
        validate: validate.slug
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
serviceCreate.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var retrieveInput;

    var name = ctx.params[0];
    if (name) {
      var errors = validator({ name: name });
      if (errors.length > 0) {
        retrieveInput = Promise.reject(errors[0]);
      } else {
        retrieveInput = Promise.resolve({
          name: name
        });
      }
    } else {
      retrieveInput = serviceCreate._prompt();
    }

    return retrieveInput.then(function (userInput) {
      return serviceCreate._execute(ctx.session, userInput);
    }).then(resolve).catch(reject);
  });
};

/**
 * Attempt to create service with supplied input
 *
 * @param {object} session - Session object
 * @param {object} userInput
 */
serviceCreate._execute = function (session, userInput) {
  return new Promise(function (resolve, reject) {
    if (!(session instanceof Session)) {
      throw new TypeError('Session object missing on Context');
    }

    client.auth(session.token);

    return client.post({
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
serviceCreate._prompt = function () {
  var prompt = new Prompt({
    stages: serviceCreate.questions
  });

  return prompt.start();
};
