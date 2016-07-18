'use strict';

var Promise = require('es6-promise').Promise;

var Prompt = require('../../cli/prompt');
var Command = require('../../cli/command');

var user = require('../../user');
var login = require('../../login');
var verify = require('../../verify');

module.exports = new Command(
  'signup',
  'create an Arigato account',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      // Create prompt from user questions
      var prompt = new Prompt({
        stages: user.questions
      });

      // Begin asking questions
      prompt.start().then(function (userInput) {
        // Create user object from input
        return user.create(ctx.api, userInput).then(function () {
          return userInput;
        });

      // Success, account created but now login
      })
      .then(function (userInput) {
        var params = {
          email: userInput.email,
          passphrase: userInput.passphrase
        };
        return login.subcommand(ctx, params);
      })
      .then(function () {
        user.output.success();

        verify.output.intermediate({ top: false });
        return verify.subcommand(ctx).then(function (result) {
          verify.output.success();
          login.output.success({ top: false });
          return result;
        });

      // Flow complete
      })
      .then(function () {
        resolve();

      // Account creation failed
      })
      .catch(function (err) {
        if (err && err.output !== false) {
          err.type = err.type || 'unknown';
          user.output.failure(err);
        }
        reject(err);
      });
    });
  }
);
