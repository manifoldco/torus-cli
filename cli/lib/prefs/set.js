'use strict';

var set = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var output = require('../cli/output');
var rc = require('./rc');

set.output = {};

set.output.success = output.create(function () {
  console.log('Preferences updated.');
});

set.output.failure = output.create(function () {
  console.log('Preferences could not be updated, please try again.');
});

set.execute = function (ctx) {
  var prefs = ctx.prefs;
  var params = {
    prefsKey: ctx.params[0],
    prefsVal: ctx.params[1] || true
  };

  if (!params.prefsKey && !_.isString(params.prefsKey)) {
    return Promise.reject(new Error('Missing required [key] parameter'));
  }

  if (params.prefsVal === 'true') {
    params.prefsVal = true;
  }

  if (params.prefsVal === 'false') {
    params.prefsVal = false;
  }

  if (!_.isBoolean(params.prefsVal) && !_.isString(params.prefsVal)) {
    return Promise.reject(new Error('Missing required [value] parameter'));
  }

  try {
    prefs.set(params.prefsKey, params.prefsVal);
  } catch (err) {
    return Promise.reject(err);
  }

  return rc.write(prefs);
};
