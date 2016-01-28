'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const apps = exports;

apps.create = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);

    var opts = {
      uri: '/apps',
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

apps.get = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);

    var opts = {
      uri: '/apps/'+params.app_id
    };

    client.get(opts).then((results) => {
      if (typeof results.body !== 'object' || !results.body.uuid) {
        return reject(new APIError(
          'Improper data returned from the registry'
        ));
      }

      resolve(results.body);
    }).catch(reject);
  });
};
