'use strict';

const hmac = exports;
const createHmac = require('crypto').createHmac;

const HMAC_ALGORITHM = 'sha512';

hmac.generate = function (key, value) {
  console.log(arguments);
  var digest = createHmac(HMAC_ALGORITHM, key);
  digest.update(value);

  return Promise.resolve(digest.digest('base64'));
};
