'use strict';

var verify = exports;

var client = require('../api/client').create();
var validate = require('../validate');

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
 * Attempt log with supplied credentials
 *
 * @param {object} daemon - Daemon object
 * @param {object} userInput
 */
verify.attempt = function(daemon, userInput) {
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
