'use strict';

var utils = exports;

var _ = require('lodash');
var crypto = require('crypto');
var Promise = require('es6-promise').Promise;

utils.randomBytes = function(length) {
  if (!_.isNumber(length) || parseInt(length, 10) < 1) {
    throw new Error('length required');
  }
  return new Promise(function(resolve, reject) {
    crypto.randomBytes(length, function(err, buf) {
      if (err) {
        return reject(err);
      }
      resolve(buf);
    });
  });
};
