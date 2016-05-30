'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var services = require('../../services');
var auth = require('../../middleware/auth');

var cmd = new Command(
  'services:info [name]',
  'view information about a service',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      services.info.execute(ctx).then(function (service) {
        services.info.output.success(null, service);

        resolve(true);
      }).catch(function (err) {
        services.info.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());

module.exports = cmd;
