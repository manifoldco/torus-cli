'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const envs = exports;

envs.create = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);

    var opts = {
      uri: '/apps/'+params.app_id+'/envs',
      json: {
        name: params.name
      }
    };

    client.post(opts).then((results) => {
      if (typeof results.body !== 'object' || !results.body.uuid) {
        return reject(new APIError(
          'Improper data returned from the registry'
        ));
      }

      resolve(results.body);
    }).catch(reject);
  });
};

envs.list = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);
    var opts = {
      uri: '/apps/'+params.app_id+'/envs'
    };

    client.get(opts).then((results) => {
      if (!results.body || !Array.isArray(results.body)) {
        return reject(new APIError());
      }

      resolve(results.body);
    }).catch(reject);
  });
};
