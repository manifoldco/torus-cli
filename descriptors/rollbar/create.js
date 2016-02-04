'use strict';

const _ = require('lodash');
const request = require('request');

const accessTokenMap = {
  'write': 'access_token_write',
  'read': 'access_token_read',
  'post_client_item': 'access_token_client',
  'post_server_item': 'access_token_server'
};

/**
 * Promisified request
 *
 * @param {object} opts - Request options
 * @param {number} statusCode - Expected status code
 */
function requestAsync(opts, statusCode) {
  return new Promise((resolve, reject) => {
    request(opts, function(err, res, body) {
      if (err) {
        return reject(err);
      }
      if (statusCode !== res.statusCode) {
        err = new Error('Unexpected status code: ' + res.statusCode);
        return reject(err);
      }

      resolve(_.isString(body)? JSON.parse(body) : body);
    });
  });
}

module.exports = function (deployment, descriptor, params) {
  return new Promise((resolve, reject) => {
    var apiUrl = deployment.get('locations').find((loc) => {
      return loc.name === 'api';
    });
    var credential = descriptor.get('credentials').find((cred) => {
      return cred.name === 'project';
    });

    if (!apiUrl) {
      return reject(new Error(
        'Could not find api url'
      ));
    }
    if (!credential) {
      return reject(new Error(
        'Could not find cred type: project'
      ));
    }

    var basePath = descriptor.get('basePath');
    var headers = { 'content-type': 'application/json' };
    var endpoint = `${apiUrl.url}${basePath}`;

    /**
     * Create new project
     */
    var createProject = {
      method: 'POST',
      url: `${endpoint}/projects`,
      json: {
        name: 'arigato_' + Date.now(),
      },
      qs: {
        access_token: params.provided.master_credentials.access_token_write
      },
      headers: headers
    };

    /**
     * Retrieve project access tokens
     */
    var getProject = {
      method: 'GET',
      qs: {
        access_token: params.provided.master_credentials.access_token_read
      },
      headers: headers
    };

    // Generate a new project for the environment
    var projectId;
    requestAsync(createProject, 200).then(function(body) {
      if (!body || !body.result || !body.result.id) {
        var err = new Error('Unexpected body: project not found');
        return reject(err);
      }

      return ''+_.get(body, 'result.id');

    // Retrieve the tokens for the project
    }).then(function(id) {
      projectId = id;
      getProject.url = `${endpoint}/project/${projectId}/access_tokens`;
      return requestAsync(getProject, 200);

    // Identify tokens from response
    }).then(function(body) {
      if (!body || !_.isArray(body.result)) {
        var err = new Error('Unexpected body: no tokens found');
        return reject(err);
      }

      // Identify and rename tokens for storage
      var tokens = {};
      body.result.forEach(function(token) {
        if (accessTokenMap[token.name]) {
          tokens[accessTokenMap[token.name]] = token.access_token;
        }
      });

      // Return comprehensive configuration
      var config = _.extend({ project_id: projectId }, tokens);
      resolve(config);
    }).catch(reject);
  });
};
