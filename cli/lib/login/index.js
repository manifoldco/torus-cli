'use strict';

var login = exports;

var Promise = require('es6-promise').Promise;
var user = require('common/crypto/user');

var Prompt = require('../cli/prompt');
var output = require('../cli/output');
var validate = require('../validate');
var Session = require('../session');

var TYPE_LOGIN = 'login';
var TYPE_AUTH = 'auth';

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
  ctx.params = [];
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
  ctx.api.reset(); // Clear any existing authorization

  var salt;
  var loginToken;
  return ctx.api.tokens.create({
    type: TYPE_LOGIN,
    email: userInput.email
  })
  .then(function (result) {
    salt = result.salt;
    loginToken = result.login_token;

    // Catch invalid data should API not return proper error status
    if (!salt || !loginToken) {
      throw new Error('invalid response from api');
    }

    return user.deriveLoginHmac(userInput.passphrase, salt, loginToken);
  })
  .then(function (loginTokenHmac) {
    // Use the login token to make an authenticated login attempt
    ctx.api.auth(loginToken);

    return ctx.api.tokens.create({
      type: TYPE_AUTH,
      login_token_hmac: loginTokenHmac
    })
    .then(function (result) { // eslint-disable-line
      // Re-authorize the api client for subsequent requests
      var authToken = result.auth_token;
      ctx.api.auth(authToken);

      var sessionData = {
        token: authToken,
        passphrase: userInput.passphrase
      };

      return ctx.daemon.set(sessionData).then(function () {
        return new Session(sessionData);
      });
    });
  });
};
