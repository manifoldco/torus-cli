'use strict';

var validate = exports;

var util = require('util');

var _ = require('lodash');
var validator = require('validator');

/**
 * TODO: Change js validation for json schema
 * https://github.com/arigatomachine/cli/issues/134
 */
function ValidationError(message, code) {
  Error.captureStackTrace(this, this.constructor);

  this.name = this.constructor.name;
  this.message = message || 'Validation Error';
  this.code = code || 'client_validation_error';
  this.type = 'validation_error';
}
util.inherits(ValidationError, Error);

validate.ValidationError = ValidationError;

/**
 * Given a map of names to validation functions it returns a function that
 * validates that all of the object data must match.
 *
 * @param {Object} ruleMap map of value names to validation functions
 * @param {Boolean} requireAll defaults to true, require all to be present
 * @returns {Function} function accepting map of value names to values which
 *                     returns an empty array on success or an array of errors.
 */
validate.build = function (ruleMap, requireAll) {
  requireAll = (requireAll === undefined) ? true : requireAll;

  return function (input) {
    var keyDiff = _.difference(Object.keys(ruleMap), Object.keys(input));
    if (keyDiff.length > 0 && requireAll) {
      return [new ValidationError('Missing parameters: ' + keyDiff.join(', '))];
    }

    var errs = _.map(ruleMap, function (rule, name) {
      if (input[name] === undefined && !requireAll) {
        return null;
      }

      var output = rule(input[name]);
      return (typeof output === 'string') ?
        new ValidationError(name + ': ' + output) : null;
    });

    return errs.filter(function (err) { return err !== null; });
  };
};

var CRED_NAME = new RegExp(/^[a-z][a-z0-9_]{0,63}$/);

validate.credName = function (input) {
  var error =
    'Credential must contain only alphanumeric or underscore characters';

  return CRED_NAME.test(input) ? true : error;
};

validate.name = function (input) {
  var error = 'Please provide your full name';
  return input.length < 3 || input.length > 64 ? error : true;
};

validate.slug = function (input) {
  var error = 'Only alphanumeric, hyphens and underscores are allowed';
  return validator.matches(
    input, /^[a-z0-9][a-z0-9\-_]{0,63}$/) ? true : error;
};

validate.email = function (input) {
  var error = 'Please enter a valid email address';
  return validator.isEmail(input) ? true : error;
};

validate.passphrase = function (input) {
  var error = 'Passphrase must be at least 8 characters';
  return input.length < 8 ? error : true;
};

validate.id = function (input) {
  var error = 'Please enter a valid 24 character ID';
  var matches = validator.matches(input, /^[a-zA-Z0-9\\_\\-]+$/);
  return input.length !== 24 || !matches ? error : true;
};

validate.code = function (input) {
  var error = 'Verification code must be exactly 9 characters';
  var trimmed = input.replace(/\s/g, '');
  return trimmed.length !== 9 ? error : true;
};
