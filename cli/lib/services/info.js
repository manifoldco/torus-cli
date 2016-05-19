'use strict';

var serviceInfo = exports;

var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var client = require('../api/client').create();

serviceInfo.output = {};

serviceInfo.output.success = output.create(function(service) {
  console.log('Service '+service.body.name+':');
  console.log('');
  console.log('Nothing here yet! Work in progress :)');
});

serviceInfo.output.failure = output.create(function() {
  console.log('Retrieval of service failed!');
});

var validator = validate.build({
  name: validate.slug
});

/**
 * Retrieve a specific service by name
 */
serviceInfo.execute = function (ctx) {
  return new Promise(function(resolve, reject) {
    var name = ctx.params[0];
    var errors = validator({ name: name });
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    if (!ctx.token) {
      return reject(new Error('must authenticate first'));
    }

    client.auth(ctx.token);
    client.get({ url: '/services/'+name }).then(function(results) {
      if (!results.body || results.body.length !== 1) {
        return reject(new Error('service not found'));
      }

      resolve(results.body[0]);
    }).catch(reject);
  });
};
