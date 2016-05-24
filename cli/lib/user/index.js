'use strict';

var user = exports;

var _ = require('lodash');
var utils = require('common/utils');

var client = require('../api/client').create();
var validate = require('../validate');
var output = require('../cli/output');

user.crypto = require('./crypto');

user.output = {};

user.output.success = output.create(function() {
  console.log('Your account has been created!');
});

user.output.failure = function(err) {
  var message = err.message;
  var messages = Array.isArray(message)? message : [message];

  output.create.call(null, function() {
    switch (err.type) {
      // TODO: Graceful re-start of prompt for invalid input
      case 'invalid_request':
        if (messages.indexOf('resource exists') > -1) {
          console.error('Email address in use, please try again');
        } else {
          console.error('Whoops! Something went wrong. Please try again');
        }
      break;
      default:
        console.error('Signup failed, please try again');
      break;
    }
  });
};

/**
 * Signup prompt questions
 */
user.questions = function() {
  var self = this;
  return [
    // Stage one
    [
      {
        name: 'name',
        message: 'Full name',
        validate: validate.name,
      },
      {
        name: 'email',
        message: 'Email',
        validate: validate.email,
      },
    ],
    // Stage two
    [
      {
        type: 'password',
        name: 'passphrase',
        message: 'Passphrase',
        validate: validate.passphrase,
        retryMessage: 'Passwords did not match. Please re-enter your password',
      },
      {
        type: 'password',
        name: 'confirm_passphrase',
        message: 'Confirm Passphrase',
        validate: function(input, answers) {
          var error = 'Passphrase must match';
          var failed = input !== answers.passphrase;
          // When failed, tell prompt which stage failed to recurse
          // if the maximum attempts has been reached
          return failed? self.failed(1, error) : true;
        },
      },
    ]
  ];
};

/**
 * Create user object from inputs
 *
 * @param {object} userInput
 */
user.create = function(userInput) {
  var object = {
    id: utils.id('user'), // generate user-object id
    body: {
      name: userInput.name,
      email: userInput.email,
    }
  };

  // Encrypt the password, generate the master key
  return user.crypto.encryptPasswordObject(userInput.passphrase)
    .then(function(result) {
      // Append the master and password objects to body
      object.body = _.extend(object.body, result);
      return object;
    }).then(function() {
      return client.post({
        url: '/users',
        json: object
      });
    });
};
