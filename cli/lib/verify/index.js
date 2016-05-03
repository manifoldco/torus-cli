'use strict';

var verify = exports;

var Promise = require('es6-promise').Promise;

var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

verify.output = {};

/**
 * Intermediate output
 */
verify.output.intermediate = function() {
  console.log('We have sent a verification code to your email.');
  console.log('Please paste your code here');
  console.log('');
};

/**
 * Success output
 */
verify.output.success = function() {
  // TODO: Proper output module for errors and banner messages
  console.log('');
  console.log('Your account is now verified');
  console.log('');
};

/**
 * Failure output
 */
verify.output.failure = function() {
  console.error('');
  console.error('Email verification failed, please try again');
  console.error('');
};

/**
 * Login prompt questions
 *
 * @param {object} ctx - Prompt context
 */
verify.questions = function(/*ctx*/) {
  return [
    [
      {
        name: 'code',
        message: 'Verification code',
        validate: validate.code,
      },
    ]
  ];
};

/**
 * Create prompt for verify
 *
 * @param {object} ctx - Command context
 */
verify.execute = function(ctx) {
  var retrieveInput;
  var code = verify._codeFromParams(ctx.params);

  // Code found as param, attempt verification
  if (code) {
    retrieveInput = Promise.resolve({
      code: code
    });

  // No code passed as param, request from user
  } else {
    retrieveInput = verify._prompt();
  }

  return retrieveInput.then(function(userInput) {
    return verify._execute(ctx.daemon, userInput);
  });
};

/**
 * Process verification
 *
 * @param {object} ctx - Command context
 */
verify.subcommand = function(ctx) {
  ctx.params = [];
  return new Promise(function(resolve, reject) {
    verify.execute(ctx).then(resolve).catch(function(err) {
      verify.output.failure();
      err.output = false;
      reject(err);
    });
  });
};

/**
 * Create prompt object
 */
verify._prompt = function() {
  var prompt = new Prompt(verify.questions);
  return prompt.start();
};

/**
 * Attempt log with supplied credentials
 *
 * @param {object} daemon - Daemon object
 * @param {object} userInput
 */
verify._execute = function(daemon, userInput) {
  return daemon.get('token').then(function(result) {
    client.auth(result.token);

    if (!client.authToken) {
      throw new Error('must authenticate first');
    }

    return client.post({
      url: '/users/verify',
      json: {
        // Trim spaces if provided in input
        code: userInput.code.replace(/\s/g, ''),
      }
    }).then(function(result) {
      return result.user;
    });
  });
};

/**
 * Identify code from params
 *
 * @param {array} params
 */
verify._codeFromParams = function(params) {
  params = params || [];
  var code;
  if (params.length === 3) {
    code = params.join('');
  } else if (params.length === 1) {
    code = params[0];
  }
  return code;
};
