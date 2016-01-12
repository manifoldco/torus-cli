'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const users = exports;

users.create = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client();

    var opts = {
      uri: '/users',
      json: {
        name: params.name,
        email: params.email,
        salt: params.salt.toString('base64'),
        password: params.password.toString('base64')
      }
    };

    client.post(opts).then((results) => {
      var body = results.body;
      if (!body || typeof body !== 'object') {
        return reject(new APIError(
          'User was not returned in the body',
          'missing_data_error'
        ));
      }

      return resolve(body);
    }).catch(reject);
  });
};

users.get = function (params) {
  return new Promise((resolve, reject) => {
    if (!params.session_token || !params.uuid) {
      return reject(new Error('Missing parameters'));
    }

    var client = new Client(null, params.sessionToken);

    var opts = {
      uri: `/users/${params.uuid}`
    };

    client.get(opts).then((results) => {
      return resolve(results.body);
    }).catch(reject);
  });
};
