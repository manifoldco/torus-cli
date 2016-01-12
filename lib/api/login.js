'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const login = exports;

login.get_session = function (email) {
  return new Promise((resolve, reject) => {
    var client = new Client();

    var opts = {
      uri: '/login/session',
      json: {
        email: email
      }
    };

    client.post(opts).then((results) => {
      if (!results.body || !results.body.salt || !results.body.login_token) {
        return reject(new APIError(
          'Login details were not return in the body',
          'missing_data_error'
        ));
      }

      resolve({
        salt: new Buffer(results.body.salt, 'base64'),
        login_token: results.body.login_token
      });
    }).catch(reject);
  });
};

login.authenticate = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.login_token);

    var opts = {
      uri: '/login',
      json: {
        password: params.pwh.toString('base64')
      }
    };

    client.post(opts).then((results) => {
      if (!results.body || !results.body.user || !results.body.session_token) {
        return reject(new APIError(
          'Login details were not returned in the body',
          'missing_data_error'
        ));
      }

      resolve(results.body);
    }).catch(reject);
  });
};
