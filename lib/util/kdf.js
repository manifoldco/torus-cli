'use strict';

const triplesec = require('triplesec');

const kdf = exports;

const SCRYPT_N = Math.pow(2, 16); // factor to control cpu/mem suage (2^16) 
const SCRYPT_R = 8; // block size factor
const SCRYPT_P = 1; // parallelism factor
const SCRYPT_DKLEN = 192; // generate 192 bit key

kdf.generate = function (key, salt, hook) {

  return new Promise((resolve, reject) => {
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
