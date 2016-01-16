'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const collaborators = exports;

collaborators.add = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);

    var opts = {
      uri: '/apps/'+params.app_id+'/collaborators',
      json: {
        user: params.email
      }
    };

    client.post(opts).then((results) => {
      if (typeof results.body !== 'object') {
        return reject(new APIError());
      }

      resolve(results.body);
    }).catch(reject);
  });
};

