'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var logout = require('../../logout');
var auth = require('../../middleware/auth');

function output(msg) {
  console.log('\n ' + msg + ' \n');
}

var cmd = new Command(
  'logout',
  'Logout of your Arigato account',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      logout(ctx)
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

cmd.hook('pre', auth());

module.exports = cmd;
