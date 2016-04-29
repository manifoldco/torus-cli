'use strict';

var user = exports;

var kdf = require('../crypto/kdf');
var utils = require('../crypto/utils');
var triplesec = require('../crypto/triplesec');

var algos = require('common/types/algos');
var base64url = require('base64url');

var SALT_BYTES = 16;
var MASTER_KEY_BYTES = 256;
var SLICE_LENGTH_BYTES = 192;
var REMAINDER_SLICE_LENGTH_BYTES = 32;

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
user.encryptPasswordObject = function(password) {
  var data = {};

  // Generate 128 bit (16 byte) salt for password
  return utils.randomBytes(SALT_BYTES)
    // Construct the password object
    .then(function(passwordSalt) {
      data.password = {
        salt: base64url.encode(passwordSalt),
        alg: algos.value('scrypt'), // 0x23
      };

      // Create password buffer
      return kdf.generate(password, passwordSalt).then(function(buf) {
        // Append the base64url value
        data.password.value = user.pwh(buf);
        return buf;
      });

    // Construct the master object from the password buffer
    }).then(function(passwordBuf) {
      data.master = {
        alg: algos.value('triplesec v3'), // 0x22
      };

      // Generate 1024 bit (256 byte) master key
      return utils.randomBytes(MASTER_KEY_BYTES).then(function(masterKeyBuf) {
        var passwordBufSlice = passwordBuf.slice(0, SLICE_LENGTH_BYTES);
        // Ensure the buffer used to encrypt is 192 bytes
        if (passwordBufSlice.length !== SLICE_LENGTH_BYTES) {
          throw new Error('invalid buffer length');
        }
        // Encrypt master key using the password buffer
        return triplesec.encrypt({
          data: masterKeyBuf,
          key: passwordBufSlice,
        }).then(function(buf) {
          // Base64 the master value for transmission
          data.master.value = base64url.encode(buf);
          return data;
        });
      });
    });
};

/**
 * Slice password and encode
 *
 * @param {buffer} passwordBuf
 */
user.pwh = function(passwordBuf) {
  // pwh is created from last 32 byte / 256 bit of password buffer
  var passwordBufSlice = passwordBuf.slice(SLICE_LENGTH_BYTES);
  if (passwordBufSlice.length !== REMAINDER_SLICE_LENGTH_BYTES) {
    throw new Error('invalid buffer length');
  }
  return base64url.encode(passwordBufSlice);
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
    var value = base64url.toBuffer(userObject.body.master.value);
    // Use the password buffer to decrypt the master key
    // 192 byte slice
    var masterKey = buf.slice(0, SLICE_LENGTH_BYTES);
    if (masterKey.length !== SLICE_LENGTH_BYTES) {
      throw new Error('invalid buffer length');
    }
    // Returns masterKey buffer for use with encrypting
    return triplesec.decrypt({
      data: value,
      key: masterKey,
    });
  });
};
