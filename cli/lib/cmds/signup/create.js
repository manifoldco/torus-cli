'use strict';

var validator = require('validator');
var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var Prompt = require('../../cli/prompt');

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

      var prompt = new Prompt(stages);
      prompt.start().then(function(result) {
        resolve(result);
      }).catch(function(err) {
        reject(err);
      });
    });
  }
);
