'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const keys = exports;

keys.upload = function (params) {
  return new Promise((resolve, reject) => {

    var client = new Client(null, params.session_token);
    var opts = {
      uri: '/keys/',
      json: {
        login_session: params.login_token,
        password: params.password.toString('base64'),
        public_key: params.public_key,
        private_key: params.private_key.toString('base64')
      }
    };

    client.post(opts).then((results) => {
      if (!results.body || !results.body.uuid || !results.body.fingerprint) {
        return reject(new APIError(
          'Missing data',
          'missing_data_error'
        ));
      }

      resolve(results.body);
    }).catch(reject);
  });
};
