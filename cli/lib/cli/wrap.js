'use strict';

var Promise = require('es6-promise').Promise;

function mutate(exitCode) {
  if (exitCode === undefined) {
    return 0;
  }

  if (typeof exitCode === 'boolean') {
    return (exitCode === true) ? 0 : 1;
  }

  if (typeof exitCode !== 'number') {
    throw new TypeError('Exit code must be undefined, boolean, or number');
  }

  if (exitCode < 0 || exitCode > 127) {
    throw new TypeError(
      'Exit code must be a positive integer between 0 and 127');
  }

  return exitCode;
}

// WHa tdo I want to do here? if it returns no result.. then exitCode is 0
// I fit returns a boolean than exitCode = (exitCode === true) ? 0 : 1
module.exports = function wrap(fn) {
  var result;
  try {
    result = fn();
  } catch (err) {
    return Promise.reject(err);
  }

  if (!(result && result.then)) {
    try {
      return Promise.resolve(mutate(result));
    } catch (err) {
      return Promise.reject(err);
    }
  }

  return result.then(mutate);
};
