'use strict';

const hmac = exports;
const createHmac = require('crypto').createHmac;

const HMAC_ALGORITHM = 'sha512';

hmac.generate = function (key, value) {
  var digest = createHmac(HMAC_ALGORITHM, key);
  digest.update(value);

  return Promise.resolve(digest.digest('base64'));
};
