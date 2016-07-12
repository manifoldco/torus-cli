'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var set = require('../../prefs/set');
var target = require('../../middleware/target');

var cmd = new Command(
  'prefs:set [key] [value]',
  'set the preference key to the value. If value is omitted, then it sets it to true',
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

cmd.hook('pre', target());

module.exports = cmd;
