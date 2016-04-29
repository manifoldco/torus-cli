'use strict';

var login = exports;

var base64url = require('base64url');

var client = require('../api/client').create();
var validate = require('../validate');
var crypto = require('../crypto');
var kdf = require('../crypto/kdf');
var user = require('../user');

/**
 * Login prompt questions
 *
 * @param {object} ctx - Prompt context
 */
login.questions = function(/*ctx*/) {
  return [
    [
      {
        name: 'email',
        message: 'Email',
        validate: validate.email,
      },
      {
        type: 'password',
        name: 'passphrase',
        message: 'Passphrase',
      }
    ]
  ];
};

/**
 * Attempt log with supplied credentials
 *
 * @param {object} daemon - Daemon object
 * @param {object} userInput
 */
login.attempt = function(daemon, userInput) {
  client.reset(); // Clear any existing authorization

  var salt;
  var loginToken;
  return client.post({
    url: '/login/session',
    json: {
      email: userInput.email
    }
  }).then(function(result) {
    // Catch invalid data should API not return proper error status
    if (!result.body.salt || !result.body.login_token) {
      throw new Error('invalid response from api');
    }
    // Derive a higher entropy password from plaintext passphrase
    salt = base64url.toBuffer(result.body.salt);
    loginToken = result.body.login_token;
    return kdf.generate(userInput.passphrase, salt)
      .then(function(buf) {
        return user.crypto.pwh(buf);
      });
  }).then(function(pwh) {
    // Generate hmac of password hash
    return crypto.utils.hmac(loginToken, pwh);
  }).then(function(result) {
    // Use the login token to make an authenticated login attempt
    client.auth(loginToken);
    return client.post({
      url: '/login',
      json: {
        login_token_hmac: base64url.encode(result)
      }
    }).then(function(result) {
      // Re-authorize the api client for subsequent requests
      var authToken = result.body.auth_token;
      client.auth(authToken);
      return daemon.set({
        token: authToken
      });
    });
  });
};
