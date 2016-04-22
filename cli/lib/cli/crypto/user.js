'use strict';

var user = exports;

var kdf = require('./kdf');
var utils = require('./utils');
var triplesec = require('./triplesec');

var algos = require('common/types/algos');
var base64url = require('common/utils/base64url');

/**
 * Generate both the password and master objects for user.body
 *
 * @param {string} password - Plaintext password value
 */
user.encryptPassword = function(password) {
  var data = {};

  // Randombytes for passworld salt
  return utils.randomBytes(128)
    // Construct the password object
    .then(function(passwordSalt) {
      data.password = {
        salt: passwordSalt,
        alg: algos.value('Scrypt'), // 0x23
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
        alg: algos.value('TripleSec v3'), // 0x22
      };

      // Generate master key random bytes
      return utils.randomBytes(256, false).then(function(masterKeyBuf) {
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
    return triplesec.decrypt({
      data: value,
      key: masterKey,
    }).then(function(masterKeyBuf) {
      // Return base64url encoded master key
      return base64url.encode(masterKeyBuf);
    });
  });
};
