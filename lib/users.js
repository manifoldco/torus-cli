'use strict';

const base64url = require('base64url');

const vault = require('./util/vault');
const saltGenerater = require('./util/salt').generate;
const kdfGenerator = require('./util/kdf').generate;
const hmacGenerator = require('./util/hmac').generate;

const errors = require('./errors');
const usersApi = require('./api/users');
const loginApi = require('./api/login');

const users = exports;

users.create = function (params) {
  return new Promise((resolve, reject) => {

    if (!params || !params.email || !params.name || !params.password) {
      return reject(new errors.RegistryError());
    }

    saltGenerater().then((salt) => {
      return kdfGenerator(params.password, salt).then((hashedPassword) => {
        return Promise.resolve({
          salt: salt,
          password: hashedPassword
        });
      });
    }).then((results) => {

      // do http requests by calling into lower level library for just http
      var payload = {
        name: params.name,
        email: params.email,
        salt: base64url(results.salt.toString('base64')),
        password: base64url(results.password.toString('base64'))
      };

      resolve(usersApi.create(payload));
    }).catch(reject);
  });
};

users.login = function (params) {
  return new Promise((resolve, reject) => {

    // Check if a session token already exists
    vault.get().then((box) => {
      if (box.get('session_token')) {
        return reject(new errors.AlreadyAuthenticatedError());
      }

      return loginApi.get_session(params.email);
    }).then((data) => {
      return kdfGenerator(params.password, data.salt).then((pwh) => {
        return hmacGenerator(pwh, data.login_token);
      }).then((hmacValue) => {
        var opts = {
          login_token: data.login_token,
          pwh: base64url(hmacValue)
        };

        return loginApi.authenticate(opts);
      }).then((results) => {
        return vault.get().then((box) => {
          box.set('session_token', results.session_token);
          return vault.save().then(() => {
            resolve(results.session_token);
          });
        });
      });
    }).catch(reject);
  });
};

users.logout = function () {
  return vault.get().then((box) => {
    var sessionToken = box.get('session_token');
    if (sessionToken) {
      return Promise.reject(new errors.NotAuthenticatedError());
    }

    box.remove('session_token');
    return box.save();
  });
};
