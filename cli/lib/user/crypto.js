'use strict';

var user = exports;

var kdf = require('../crypto/kdf');
var utils = require('../crypto/utils');
var triplesec = require('../crypto/triplesec');

var algos = require('common/types/algos');
var base64url = require('common/utils/base64url');

/**
 * Generate both the password and master objects for user.body
 *
 * A 128bit salt is generated / used to encrypt the plaintext password
 * The encrypted password buffer is sliced [192] and encoded in base64
 * A 1024bit key is generated as used as the master key
 * The encrypted password buffer is sliced [0,192] / used to encrypt master key
 * The encrypted+encoded values are stored on the server
 *
 * @param {string} password - Plaintext password value
 */
user.encryptPassword = function(password) {
  var data = {};

  // Generate 128 bit (16 byte) salt for password
  return utils.randomBytes(16)
    // Construct the password object
    .then(function(passwordSalt) {
      data.password = {
        salt: base64url.encode(passwordSalt),
        alg: algos.value('scrypt'), // 0x23
      };
      // Create password buffer
      return kdf.generate(password, passwordSalt).then(function(buf) {
        // Append the base64url value
        data.password.value = base64url.encode(buf.slice(192));
        return buf;
      });

    // Construct the master object from the password buffer
    }).then(function(passwordBuf) {
      data.master = {
        alg: algos.value('triplesec v3'), // 0x22
      };

      // Generate 1024 bit (256 byte) master key
      return utils.randomBytes(256).then(function(masterKeyBuf) {
        // Encrypt master key using the password buffer
        return triplesec.encrypt({
          data: masterKeyBuf,
          key: passwordBuf.slice(0,192),
        }).then(function(buf) {
          // Base64 the master value for transmission
          data.master.value = base64url.encode(buf);
          return data;
        });
      });
    });
};

/**
 * Decrypt the master key from user object using password
 *
 * @param {string} password - Plaintext password
 * @param {object} userObject - Full user object
 */
user.decryptMasterKey = function(password, userObject) {
  // Use stored password salt to encrypt password
  var passwordSalt = userObject.body.password.salt;
  return kdf.generate(password, passwordSalt).then(function(buf) {
    // Decode master value from base64url
    var value = base64url.decode(userObject.body.master.value);
    // Use the password buffer to decrypt the master key
    var masterKey = buf.slice(0, 192);
    // Returns masterKey buffer for use with encrypting
    return triplesec.decrypt({
      data: value,
      key: masterKey,
    });
  });
};
