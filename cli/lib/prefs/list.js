var list = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var output = require('../cli/output');
var ini = require('ini');

list.output = {};

list.output.success = output.create(function (prefs) {
  var msg = '  ' + prefs.path;
  msg += '\n\n';
  msg += !_.isEmpty(prefs.values) ? ini.stringify(prefs.values) : 'No preferences configured.\n';

  msg = msg.split('\n').join('\n  ');

  console.log(msg);
});

list.output.failure = output.create(function () {
  console.log('Could not retrieve preferences, please try again.');
});

list.execute = function (ctx) {
  return new Promise(function (resolve) {
    return resolve(ctx.prefs);
  });
};
