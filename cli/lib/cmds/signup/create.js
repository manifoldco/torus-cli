'use strict';

var Promise = require('es6-promise').Promise;

var Prompt = require('../../cli/prompt');
var Command = require('../../cli/command');

var user = require('../../cli/user');

module.exports = new Command(
  'signup',
  'create an arigato account',
  function() {
    return new Promise(function(resolve, reject) {

      // Create prompt from user questions
      var prompt = new Prompt(user.questions);

      // Begin asking questions
      prompt.start().then(function(userInput) {

        // Create user object from input
        return user.create(userInput);

      // Success, account created
      }).then(function() {
        // TODO: Proper output module for errors and banner messages
        console.log('');
        console.log('Your account has been created!');
        console.log('');
        resolve();

      // Account creation failed
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
