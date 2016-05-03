'use strict';

var Promise = require('es6-promise').Promise;

var Prompt = require('../../cli/prompt');
var Command = require('../../cli/command');

var verify = require('../../verify');

module.exports = new Command(
  'verify',
  'verify your account\'s email address',
  function(ctx) {
    return new Promise(function(resolve, reject) {
      var params = ctx.params;

      var code;
      if (params.length === 3) {
        code = params.join('');
      } else if (params.length === 1) {
        code = params[0];
      }

      // Create prompt from login questions
      var prompt = new Prompt(verify.questions);

      var verification;
      if (code) {
        // Resolve with the known code
        verification = Promise.resolve({
          code: code
        });
      } else {
        // Ask the user for input
        verification = prompt.start();
      }

      verification.then(function(userInput) {

        // Attempt login from user input
        return verify.attempt(ctx.daemon, userInput);

      // Success, session created
      }).then(function() {
        // TODO: Proper output module for errors and banner messages
        console.log('');
        console.log('Your account is now verified');
        console.log('');
        resolve();

      // Account creation failed
      }).catch(function(err) {
        err.type = err.type || 'unknown';
        console.error('');
        console.error('Email verification failed, please try again');
        console.error('');
        reject(err);
      });
    });
  }
);
