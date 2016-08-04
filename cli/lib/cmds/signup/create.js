'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var user = require('../../user');
var login = require('../../login');
var verify = require('../../verify');

module.exports = new Command(
  'signup',
  'create an Arigato account',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      user.execute(ctx).then(function (userInput) {
        var params = {
          email: userInput.email,
          passphrase: userInput.passphrase
        };
        return login.subcommand(ctx, params);
      })
      .then(function () {
        return user.finalize(ctx);
      })
      .then(function () {
        user.output.success();

        verify.output.intermediate({ top: false });
        return verify.subcommand(ctx).then(function (result) {
          verify.output.success();
          login.output.success({ top: false });
          return result;
        });
      })
      // Flow complete
      .then(resolve)
      // Account creation failed
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
