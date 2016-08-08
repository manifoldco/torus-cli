'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var list = require('../../prefs/list');

var cmd = new Command(
  'prefs',
  'Show your account preferences',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      list.execute(ctx).then(function (payload) {
        list.output.success({ bottom: false }, payload);

        resolve(true);
      }).catch(function (err) {
        list.output.failure();
        reject(err);
      });
    });
  }
);

module.exports = cmd;
