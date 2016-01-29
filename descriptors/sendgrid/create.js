'use strict';

const _ = require('lodash');
const request = require('request');

module.exports = function (deployment, descriptor, params) {
  return new Promise((resolve, reject) => {
    var apiUrl = deployment.get('locations').find((loc) => {
      return loc.name === 'api';
    });
    var credential = descriptor.get('credentials').find((cred) => {
      return cred.name === 'api_key';
    });

    if (!apiUrl) {
      return reject(new Error(
        'Could not find api url'
      ));
    }

    if (!credential) {
      return reject(new Error(
        'Could not find cred type: api_key'
      ));
    }

    var url = apiUrl.url + '/' + descriptor.get('basePath') + '/api_keys';
    var opts = {
      url: url,
      auth: {
        username: params.provided.master_credentials.username,
        password: params.provided.master_credentials.password,
        sendImmediately: true
      },
      json: {
        name: 'arigato:'+params.app.uuid+':'+params.env.uuid,
        scopes: credential.permissions
      }
    };

    request.post(opts, function (err, res, body) {
      if (err) {
        return reject(err);
      }

      if (res.statusCode !== 201) {
        return reject(new Error('Non 201 status code: '+res.statusCode));
      }

      // Pare the object down to just what the output schema wants
      resolve(_.pick(body, Object.keys(credential.schema.properties)));
    });
  });
};
