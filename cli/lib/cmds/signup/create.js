'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var user = require('../../user');
var login = require('../../login');
var verify = require('../../verify');

var signup = new Command(
  'signup <email> <code>',
  'Create a new Arigato account which, while in alpha, requires an invite code',
  'jeff@example.com 1a2b3c4d5e\n\n Jeff signs up with an invite code he received',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      var usr;
      user.execute(ctx, ctx.params).then(function (res) {
        var params = {
          email: res.inputs.email,
          passphrase: res.inputs.passphrase
        };
        usr = res.user;
        return login.subcommand(ctx, params);
      })
      .then(function () {
        return user.finalize(ctx);
      })
      .then(function () {
        user.output.success();

        if (usr.body.state === 'active') {
          return Promise.resolve();
        }

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

module.exports = signup;
