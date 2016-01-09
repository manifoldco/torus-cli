'use strict';

const base64url = require('base64url');

const saltGenerator = require('../util/salt').generate;
const kdfGenerator = require('../util/kdf').generate;

const login = exports;

login.get_session = function (email) {
  return new Promise((resolve, reject) => {
    saltGenerator().then((salt) => {
      return kdfGenerator(salt, email).then((hash) => {
        return Promise.resolve({
          salt: base64url(salt.toString('base64')),
          login_token: base64url(hash.toString('base64')) 
        });
      });
    }).then(resolve).catch(reject);
  });
};

login.authenticate = function () {
  return new Promise((resolve, reject) => {
    saltGenerator().then((salt) => {
      resolve({
        session_token: salt
      });
    }).catch(reject);
  });
};
