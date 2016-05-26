'use strict';

var verify = exports;

var Promise = require('es6-promise').Promise;

var Session = require('../session');
var output = require('../cli/output');
var validate = require('../validate');
var Prompt = require('../cli/prompt');
var client = require('../api/client').create();

verify.output = {};

/**
 * Intermediate output
 */
verify.output.intermediate = output.create(function () {
  console.log('We have sent a verification code to your email.');
  console.log('Please enter your code below:');
});

/**
 * Success output
 */
verify.output.success = output.create(function () {
  console.log('Your account is now verified.');
});

/**
 * Failure output
 */
verify.output.failure = output.create(function () {
  console.error('Email verification failed, please try again.');
});

/**
 * Verify prompt questions
 *
 * @param {object} ctx - Prompt context
 */
verify.questions = function () {
  return [
    [
      {
        name: 'code',
        message: 'Verification code',
        validate: validate.code
      }
    ]
  ];
};

/**
 * Create prompt for verify
 *
 * @param {object} ctx - Command context
 */
verify.execute = function (ctx) {
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

  return retrieveInput.then(function (userInput) {
    return verify._execute(ctx.session, userInput);
  });
};

/**
 * Process verification
 *
 * @param {object} ctx - Command context
 */
verify.subcommand = function (ctx) {
  ctx.params = [];
  return new Promise(function (resolve, reject) {
    verify.execute(ctx).then(resolve).catch(function (err) {
      verify.output.failure();
      err.output = false;
      reject(err);
    });
  });
};

/**
 * Create prompt object
 */
verify._prompt = function () {
  var prompt = new Prompt({
    stages: verify.questions
  });
  return prompt.start();
};

/**
 * Attempt log with supplied credentials
 *
 * @param {object} session - Session object
 * @param {object} userInput
 */
verify._execute = function (session, userInput) {
  return new Promise(function (resolve, reject) {
    if (!(session instanceof Session)) {
      throw new TypeError('Session is not on Context object');
    }

    client.auth(session.token);
    client.post({
      url: '/users/verify',
      json: {
        // Trim spaces if provided in input
        code: userInput.code.replace(/\s/g, '').toUpperCase()
      }
    }).then(function (result) {
      resolve(result.user);
    }).catch(reject);
  });
};

/**
 * Identify code from params
 *
 * @param {array} params
 */
verify._codeFromParams = function (params) {
  params = params || [];
  var code;
  if (params.length === 3) {
    code = params.join('');
  } else if (params.length === 1) {
    code = params[0];
  }
  return code;
};
