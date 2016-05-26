'use strict';

function Session(obj) {
  if (!obj || typeof obj.token !== 'string' ||
      typeof obj.passphrase !== 'string') {
    throw new TypeError('Invalid object passed to create session');
  }

  this.token = obj.token;
  this.passphrase = obj.passphrase;
}

module.exports = Session;
