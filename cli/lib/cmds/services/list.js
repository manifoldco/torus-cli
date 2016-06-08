'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var services = require('../../services');
var auth = require('../../middleware/auth');

var cmd = new Command(
  'services',
  'list services for your org',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      services.list.execute(ctx).then(function (payload) {
        services.list.output.success(null, payload.body);

        resolve(true);
      }).catch(function (err) {
        services.list.output.failure();
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
