'use strict';

var Promise = require('es6-promise').Promise;
var Config = require('../config');
var rc = require('../prefs/rc');
var Prefs = require('../prefs');

module.exports = function () {
  return function (ctx) {
    var config = ctx.config;

    if (!(config instanceof Config)) {
      return Promise.reject(new Error('Must have config property on Context'));
    }

    return rc.read(config.rcPath).then(function (contents) {
      ctx.prefs = new Prefs(config.rcPath, contents);
    });
  };
};
