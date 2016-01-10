'use strict';

const uuid = require('uuid');
const base64url = require('base64url');

const saltGenerator = require('../util/salt').generate;
const kdfGenerator = require('../util/kdf').generate;

const login = exports;

login.get_session = function (email) {
  return new Promise((resolve, reject) => {
    saltGenerator().then((salt) => {
      return kdfGenerator(salt, email).then((hash) => {
        resolve({
          salt: base64url(salt.toString('base64')),
          login_token: base64url(hash.toString('base64')) 
        });
      });
    }).catch(reject);
  });
};

login.authenticate = function () {
  return new Promise((resolve, reject) => {
    saltGenerator().then((salt) => {
      resolve({
        session_token: base64url(salt.toString('base64')),
        user: {
          uuid: uuid.v4()
        }
      });
    }).catch(reject);
  });
};
