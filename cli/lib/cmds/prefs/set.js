'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var set = require('../../prefs/set');

var cmd = new Command(
  'prefs:set [key] [value]',
  'Set the preference key to the value. If value is omitted, then it sets it to true',
  'default.environment development\n\n  Sets the default environment to \'development\'',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      set.execute(ctx).then(function (payload) {
        set.output.success(null, payload);
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        set.output.failure(err);
        reject(err);
      });
    });
  }
);

module.exports = cmd;
