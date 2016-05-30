'use strict';

var login = exports;

var base64url = require('base64url');
var Promise = require('es6-promise').Promise;

var Prompt = require('../cli/prompt');
var output = require('../cli/output');
var client = require('../api/client').create();
var validate = require('../validate');
var crypto = require('../crypto');
var kdf = require('../crypto/kdf');
var user = require('../user');
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
  client.reset(); // Clear any existing authorization

  var salt;
  var loginToken;
  return client.post({
    url: '/session',
    json: {
      type: TYPE_LOGIN,
      email: userInput.email
    }
  })
  .then(function (result) {
    // Catch invalid data should API not return proper error status
    if (!result.body.salt || !result.body.login_token) {
      throw new Error('invalid response from api');
    }
    // Derive a higher entropy password from plaintext passphrase
    salt = base64url.toBuffer(result.body.salt);
    loginToken = result.body.login_token;
    return kdf.generate(userInput.passphrase, salt)
      .then(function (buf) {
        return user.crypto.pwh(buf);
      });
  })
  .then(function (pwh) {
    // Generate hmac of login token
    return crypto.utils.hmac(loginToken, pwh);
  })
  .then(function (result) {
    // Use the login token to make an authenticated login attempt
    client.auth(loginToken);
    return client.post({
      url: '/session',
      json: {
        type: TYPE_AUTH,
        login_token_hmac: base64url.encode(result)
      }
    }).then(function (result) { // eslint-disable-line
      // Re-authorize the api client for subsequent requests
      var authToken = result.body.auth_token;
      client.auth(authToken);

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
