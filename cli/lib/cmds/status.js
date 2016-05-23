'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');
var status = require('../user/status');

module.exports = new Command(
  'status',
  'shows your current cli status',
  function (ctx) {
    return new Promise(function(resolve, reject) {
      status.execute(ctx).then(function(identity) {
        status.output.success(null, identity);
        resolve();
      }).catch(function(err) {
        status.output.failure(err);
        reject(err);
      });
    });
  }
);
