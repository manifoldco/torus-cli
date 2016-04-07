'use strict';

var Promise = require('es6-promise').Promise;

module.exports = function wrap (fn) {
  try {
    var result = fn();
  } catch (err) {
    return Promise.reject(err);
  }

  if (!(result && result.then)) {
    return Promise.resolve(result); 
  }

  return result;
};
