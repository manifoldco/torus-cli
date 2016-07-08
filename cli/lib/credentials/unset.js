'use strict';

var Promise = require('es6-promise').Promise;

var credentials = require('./credentials');
var cValue = require('./value');
var harvest = require('./harvest');

var unset = exports;
unset.output = {};

unset.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (ctx.params.length < 1) {
      return reject(new Error('You must provide one parameter'));
    }

    // to unset is just setting the cred as undefined
    var value = cValue.create(undefined);
    var params = harvest.create(ctx);

    return credentials.create(ctx.api, params, value)
      .then(resolve).catch(reject);
  });
};

unset.output.success = function (ctx, cred) {
  console.log('Variable ' + cred.body.name + ' has been unset at ' +
              cred.body.pathexp + '/' + cred.body.name + '.\n');
};

unset.output.failure = function () {
  console.log('Failed to set credential value');
};
