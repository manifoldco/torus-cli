'use strict';

var Promise = require('es6-promise').Promise;

var validate = require('../validate');
var cValue = require('./value');
var credentials = require('./credentials');

var view = exports;
view.output = {};

var validator = validate.build({
  environment: validate.slug,
  service: validate.slug,
  instance: validate.slug
});

view.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var serviceName = ctx.option('service').value;
    var envName = ctx.option('environment').value;
    var instance = ctx.option('instance').value;

    if (!serviceName) {
      return reject(new Error('You must provide a --service flag'));
    }

    if (!envName) {
      return reject(new Error('You must provide a --environment flag'));
    }

    var errors = validator({
      service: serviceName,
      environment: envName,
      instance: instance
    });
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    return credentials.get(ctx.session, {
      service: serviceName,
      environment: envName,
      instance: instance
    }).then(resolve).catch(reject);
  });
};

view.output.success = function (creds) {
  creds.forEach(function (cred) {
    var value = cValue.parse(cred.body.value);

    if (value.body.type === 'undefined') {
      return;
    }

    console.log(cred.body.name.toUpperCase() + '=' + value.body.value);
  });
};

view.output.failure = function () {
  console.log('It didnt work :(');
};
