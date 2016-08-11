'use strict';

var Promise = require('es6-promise').Promise;

var cValue = require('./value');
var credentials = require('./credentials');
var harvest = require('./harvest');

var view = exports;
view.output = {};

view.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var params = harvest.get(ctx);

    return credentials.get(ctx.api, params)
    .then(resolve)
    .catch(reject);
  });
};

view.output.success = function (ctx, results) {
  var creds = results.credentials;
  var isVerbose = ctx.option('verbose').value === true;

  if (isVerbose) {
    console.log('Execution Context: ' + results.path);
  }

  creds.forEach(function (cred) {
    var value = cValue.parse(cred.body.value);

    if (value.body.type === 'undefined') {
      return;
    }

    var line = cred.body.name.toUpperCase() + '=' + value.body.value;

    if (isVerbose) {
      line += ' (sourced from: ' + cred.body.pathexp + '/' +
        cred.body.name + ')';
    }

    console.log(line);
  });
};

view.output.failure = function () {
  console.log('It didnt work :(');
};
