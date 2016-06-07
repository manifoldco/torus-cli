'use strict';

var Promise = require('es6-promise').Promise;
var Command = require('../cli/command');

var viewCred = require('../credentials/view');
var auth = require('../middleware/auth');

var view = new Command(
  'view',
  'view credentials for the current service and environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      viewCred.execute(ctx).then(function (creds) {
        viewCred.output.success(creds);
        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        viewCred.output.failure(err);
        reject(err);
      });
    });
  }
);

view.hook('pre', auth());

view.option(
  '-s, --service [service]',
  'the service you are viewing credentials for',
  undefined
);

view.option(
  '-e, --environment [environment]',
  'the environment you are viewing credentials for',
  undefined
);

view.option(
  '-i, --instance [name]',
  'the instance of the service you are viewing credentials for',
  '1'
);

module.exports = view;
