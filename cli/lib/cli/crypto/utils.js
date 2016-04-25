'use strict';

var utils = exports;

var crypto = require('crypto');
var base64url = require('common/utils/base64url');
var Promise = require('es6-promise').Promise;

utils.randomBytes = function(length, encode) {
  if (!length) {
    throw new Error('length required');
  }
  encode = (typeof encode === 'undefined')? true : !!encode;
  return new Promise(function(resolve, reject) {
    crypto.randomBytes(length, function(err, buf) {
      if (err) {
        return reject(err);
      }
      resolve(encode? base64url.encode(buf) : buf);
    });
  });
};
