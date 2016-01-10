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

    users.loggedIn().then((loggedIn) => {
      if (loggedIn) {
        return reject(new errors.AlreadyAuthenticatedError());
      }

      return loginApi.get_session(params.email);
    }).then((data) => {
      var pwhData = {
        salt: data.salt,
        login_token: data.login_token,
        password: params.password
      };

      return users._calculatePwh(pwhData)
      .then(loginApi.authenticate)
      .then((results) => {
        return vault.get().then((box) => {
          box.set('session_token', results.session_token);
          box.set('uuid', results.user.uuid);

          return box.save().then(() => {
            resolve(results);
          });
        });
      });
    }).catch(reject);
  });
};

// TODO: Break this into a lib/login or something
users.sudo = function (params) {
  return new Promise((resolve, reject) => {
    loginApi.get_session(params.email).then((results) => {

      var opts = {
        password: params.password,
        salt: results.salt,
        login_token: results.login_token
      };

      return users._calculatePwh(opts).then((pwh) => {
        resolve({
          salt: results.salt,
          login_token: results.login_token,
          pwh: pwh
        });
      });
    }).catch(reject);
  });
};

users.loggedIn = function () {
  return new Promise((resolve, reject) => {
    return vault.get().then((box) => {
      return resolve(box.get('session_token') || box.get('uuid'));
    }).catch(reject);
  });
};

users._calculatePwh = function (params) {
  return new Promise((resolve, reject) => {
    return kdfGenerator(params.password, params.salt).then((pwh) => {
      return hmacGenerator(pwh, params.login_token);
    }).then((hmacValue) => {
      resolve({
        login_token: params.login_token,
        pwh: hmacValue
      });
    }).catch(reject);
  });
};

users.logout = function () {
  return new Promise((resolve, reject) => {
    users.loggedIn().then((loggedIn) => {
      if (!loggedIn) {
        return reject(new errors.NotAuthenticatedError());
      }

      return vault.get();
    }).then((box) => {
      box.clear();
      return box.save();
    }).then(resolve).catch(reject);
  });
};
