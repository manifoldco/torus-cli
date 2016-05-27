'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var logout = require('../../logout');

function output(msg) {
  console.log('\n ' + msg + ' \n');
}

module.exports = new Command(
  'logout',
  'logout of your Arigato account',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      if (!ctx.token) {
        output('You must be logged-in, to log-out.');
        return resolve();
      }

      return logout(ctx)
        .then(function () {
          output('You have successfully logged-out. o/');
          resolve();
        }).catch(function (err) {
          err.type = err.type || 'unknown';
          output('Logout failed, please try again.');
          reject();
        });
    });
  }
);
