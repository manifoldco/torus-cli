'use strict';

var triplesec = require('triplesec');
var Promise = require('es6-promise').Promise;

var kdf = exports;

var SCRYPT_N = Math.pow(2, 15); // factor to control cpu/mem suage (2^15)
var SCRYPT_R = 8; // block size factor
var SCRYPT_P = 1; // parallelism factor
var SCRYPT_DKLEN = 224; // generate 224 byte key

kdf.generate = function (key, salt, hook) {
  return new Promise(function(resolve, reject) {
    key = (Buffer.isBuffer(key) ? key : new Buffer(key));
    salt = (Buffer.isBuffer(salt)) ? salt : new Buffer((salt));

    var params = {
      N: SCRYPT_N,
      r: SCRYPT_R,
      p: SCRYPT_P,
      dkLen: SCRYPT_DKLEN,
      key: triplesec.WordArray.alloc(key),
      salt: triplesec.WordArray.alloc(salt),
      progress_hook: hook
    };

    triplesec.scrypt(params, function (res) {
      if (res instanceof Error) {
        return reject(res);
      }

      resolve(res.to_buffer());
    });
  });
};
