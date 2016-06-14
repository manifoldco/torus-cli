'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var services = require('../../services');
var auth = require('../../middleware/auth');

var cmd = new Command(
  'services:create [name]',
  'create a new service for your org',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      services.create.execute(ctx).then(function () {
        services.create.output.success();
        resolve();

      // Service creation failed
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        services.create.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.option(
  '-o, --org [org]',
  'Specify an organization',
  null
);

cmd.hook('pre', auth());

module.exports = cmd;
