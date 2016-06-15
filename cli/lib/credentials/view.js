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

    return credentials.get(ctx.session, params)
    .then(resolve)
    .catch(reject);
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
