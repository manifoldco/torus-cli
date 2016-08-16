'use strict';

var login = exports;

var Promise = require('es6-promise').Promise;

var Prompt = require('../cli/prompt');
var output = require('../cli/output');
var validate = require('../validate');

login.output = {};

login.output.success = output.create(function () {
  console.log('You are now authenticated.');
});

login.output.failure = output.create(function () {
  console.error('Login failed, please try again.');
});

/**
 * Login prompt questions
 */
login.questions = function () {
  return [
    [
      {
        name: 'email',
        message: 'Email',
        validate: validate.email
      },
      {
        type: 'password',
        name: 'passphrase',
        message: 'Passphrase'
      }
    ]
  ];
};

/**
 * Create prompt for login
 *
 * @param {object} ctx - Command context
 * @param {object} inputs - Optional user inputs
 */
login.execute = function (ctx, inputs) {
  var retrieveInput;
  if (inputs) {
    retrieveInput = Promise.resolve(inputs);
  } else {
    retrieveInput = login._prompt();
  }

  return retrieveInput.then(function (userInput) {
    return login._execute(ctx, userInput);
  });
};

/**
 * Process login
 *
 * @param {object} ctx - Command context
 * @param {object} inputs - Optional user inputs
 */
login.subcommand = function (ctx, inputs) {
  return new Promise(function (resolve, reject) {
    login.execute(ctx, inputs).then(resolve).catch(function (err) {
      login.output.failure();
      err.output = false;
      reject(err);
    });
  });
};

/**
 * Create prompt object
 */
login._prompt = function () {
  var prompt = new Prompt({
    stages: login.questions
  });
  return prompt.start();
};

/**
 * Attempt log with supplied credentials
 *
 * @param {object} ctx - Context object
 * @param {object} userInput
 */
login._execute = function (ctx, userInput) {
  return ctx.api.login.post({
    email: userInput.email,
    passphrase: userInput.passphrase
  });
};
