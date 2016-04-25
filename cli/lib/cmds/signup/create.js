'use strict';

var _ = require('lodash');
var validator = require('validator');
var Promise = require('es6-promise').Promise;
var utils = require('common/utils');

var Prompt = require('../../cli/prompt');
var Command = require('../../cli/command');
var userCrypto = require('../../cli/crypto/user');

var DEFAULT_USER_STATE = 'unverified';

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
        // Basic user information for object body
        object.body = _.extend(object.body, {
          name: userInput.name,
          email: userInput.email,
          state: DEFAULT_USER_STATE
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
      }).then(resolve).catch(reject);
    });
  }
);
