'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var verify = require('../../verify');
var auth = require('../../middleware/auth');

var cmd = new Command(
  'verify',
  'verify your account\'s email address',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      verify.execute(ctx).then(function () {
        verify.output.success();
        resolve();

      // Account creation failed
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        verify.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());

module.exports = cmd;
