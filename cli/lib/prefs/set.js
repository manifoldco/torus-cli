'use strict';

var set = exports;

var _ = require('lodash');
var path = require('path');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var rc = require('./rc');
var keyFile = require('../util/keyfile');

var validation = {
  'core.public_key_file': keyFile.validate
};

var preparation = {
  'core.public_key_file': function (keyFilePath) {
    return Promise.resolve(path.resolve(process.cwd(), keyFilePath));
  }
};

set.output = {};

set.output.success = output.create(function () {
  console.log('Preferences updated.');
});

set.output.failure = output.create(function () {
  console.log('Preferences could not be updated, please try again.');
});

set.execute = function (ctx) {
  var prefs = ctx.prefs;

  var key = ctx.params[0];
  var value = ctx.params[1] || true;

  if (!key && !_.isString(key)) {
    return Promise.reject(new Error('Missing required [key] parameter'));
  }

  key = key.toLowerCase();

  if (value === 'true') {
    value = true;
  }

  if (value === 'false') {
    value = false;
  }

  if (!_.isBoolean(value) && !_.isString(value)) {
    return Promise.reject(new Error('Missing required [value] parameter'));
  }

  var validator = validation[key] ?
    validation[key](value) : Promise.resolve(true);

  return validator.then(function (validationStr) {
    if (_.isString(validationStr)) {
      throw new Error(key + ' is not valid: ' + validationStr);
    }

    var preparator = preparation[key] ?
      preparation[key](value) : Promise.resolve(value);

    return preparator.then(function (preparedValue) {
      prefs.set(key, preparedValue);
      return rc.write(prefs);
    });
  });
};
