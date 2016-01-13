'use strict';

const APIError = require('./client').APIError;
const Client = require('./client').Client;

const credentials = exports;

credentials.create = function (params) {
  return new Promise((resolve, reject) => {

    if (['app','env'].indexOf(params.type) === -1) {
      return reject(new TypeError('Invalid type of credential:'+params.type));
    }

    var client = new Client(null, params.session_token);

    var opts = {
      uri: '/credentials',
      json: {
        owner: params.owner,
        type: params.type,
        name: params.name,
        value: params.value
      }
    };

    client.post(opts).then((results) => {
      if (typeof results.body !== 'object') {
        return reject(new APIError());
      }

      resolve(results.body);
    });
  });
};

credentials.list = function (params) {
  return new Promise((resolve, reject) => {
    
    if (['app','env'].indexOf(params.type) === -1) {
      return reject(new TypeError('Invalid type of credential:'+params.type));
    }

    var client = new Client(null, params.session_token);
    var opts = {
      uri: '/credentials',
      qs: {}
    };

    opts.qs[params.type] = params.owner;

    client.get(opts).then((results) => {
      if (!Array.isArray(results.body)) {
        return reject(new APIError());
      }

      resolve(results.body);
    }).catch(reject);
  });
};
