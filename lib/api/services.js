'use strict';

var APIError = require('./client').APIError;
var Client = require('./client').Client;

const services = exports;

services.add = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);

    var opts = {
      uri: '/apps/'+params.app.uuid+'/services',
      method: 'post',
      json: {
        name: params.name,
        service: params.service
      }
    };

    client.post(opts).then((results) => {
      if (typeof results.body !== 'object') {
        return reject(new APIError(
          'Invalid body returned from Registry'
        ));
      } 

      resolve(results.body);
    }).catch(reject);
  });
}; 

services.get = function (params) {
  return new Promise((resolve, reject) => {
    var client = new Client(null, params.session_token);
    
    var opts = {
      uri: '/apps/'+params.app.uuid+'/services',
      method: 'get'
    };

    return client.get(opts).then((results) => {
      if (!Array.isArray(results.body)) {
        return reject(new APIError(
          'Invalid body returned from Registry'
        ));
      }

      resolve(results.body);
    }).catch(reject);
  });
};

services.supported = function () {
  return new Promise((resolve, reject) => {
    var client = new Client();
    var opts = {
      uri: '/services'
    };

    client.get(opts).then((results) => {
      if (!Array.isArray(results.body)) {
        return reject(new APIError(
          'Received invalid data from the Registry'
        ));
      }

      resolve(results.body);
    }).catch(reject);
  });
};
