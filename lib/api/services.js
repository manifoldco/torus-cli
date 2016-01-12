'use strict';

var APIError = require('./client').APIClient;
var Client = require('./client').Client;

const services = exports;

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
