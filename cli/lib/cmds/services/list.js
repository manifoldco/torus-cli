'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var services = require('../../services');

module.exports = new Command(
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
