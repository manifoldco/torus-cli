'use strict';

var _ = require('lodash');
var validator = require('validator');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Client = require('../../api/client');
var Prompt = require('../../cli/prompt');
var Command = require('../../cli/command');
var userCrypto = require('../../cli/crypto/user');

var client = new Client();

module.exports = new Command(
  'signup',
  'create an arigato account',
  function() {
    return new Promise(function(resolve, reject) {
      /**
       * Question stages
       */
      var stages = function (ctx) {
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

      var object = {
        id: utils.id('user') // generate user-object id
      };

      var prompt = new Prompt(stages);
      prompt.start().then(function(userInput) {
        // TODO: Should be created with user class
        // Issue: https://github.com/arigatomachine/cli/issues/124
        object.body = _.extend(object.body, {
          name: userInput.name,
          email: userInput.email,
        });
        return userInput;
      }).then(function(userInput) {
        // Encrypt the password, generate the master key
        return userCrypto.encryptPassword(userInput.passphrase)
          .then(function(result) {
            // Append the master and password objects to body
            object.body = _.extend(object.body, result);
            return object;
          });
      }).then(function() {
        return client.post({
          url: '/users',
          json: object
        });
      }).then(function() {
        // TODO: Proper output module for errors and banner messages
        console.log('');
        console.log('Your account has been created!');
        console.log('');
        resolve();
      }).catch(function(err) {
        err.type = err.type || 'unknown';

        var message = err.message;
        var messages = Array.isArray(message)? message : [message];

        console.error('');
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
        console.error('');

        reject(err);
      });
    });
  }
);
