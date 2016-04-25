'use strict';

var user = exports;

var _ = require('lodash');
var utils = require('common/utils');
var validator = require('validator');

var client = require('../api/client').create();

user.crypto = require('./crypto');

/**
 * Signup prompt questions
 *
 * @param {object} ctx - Prompt context
 */
user.questions = function(ctx) {
  return [
    // Stage one
    [
      {
        name: 'name',
        message: 'Full name',
        validate: function(input) {
          /**
           * TODO: Change js validation for json schema
           */
          var error = 'Please provide your full name';
          return input.length < 3 || input.length > 64? error : true;
        }
      },
      {
        name: 'email',
        message: 'Email',
        validate: function(input) {
          var error = 'Please enter a valid email address';
          return validator.isEmail(input)? true : error;
        },
      },
    ],
    // Stage two
    [
      {
        type: 'password',
        name: 'passphrase',
        message: 'Passphrase',
        validate: function(input) {
          var error = 'Passphrase must be at least 8 characters';
          return input.length < 8? error : true;
        },
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
          return failed? ctx.failed(1, error) : true;
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
  return user.crypto.encryptPassword(userInput.passphrase)
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
