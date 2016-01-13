'use strict';

var credentialsApi = require('./api/credentials');
var envsApi = require('./api/envs');
var vault = require('./util/vault');
var sendgrid = require('./integrations/sendgrid');

var credentials = exports;

credentials.initialize = function (params) {
  return new Promise((resolve, reject) => {
    return Promise.all([
      credentialsApi.create({
        session_token: params.session_token,
        type: 'app',
        owner: params.app_id,
        name: 'username',
        value: params.master_credentials.username
      }),
      credentialsApi.create({
        session_token: params.session_token,
        type: 'app',
        owner: params.app_id,
        name: 'password',
        value: params.master_credentials.password
      }),
      envsApi.list({
        session_token: params.session_token,
        app_id: params.app_id 
      })
    ]).then((results) => {
      var envs = results[2];

      // TODO: turn these into sendgrid creds
      var acquireKeys = envs.map((env) => {
        return new Promise((resolve, reject) => {
          sendgrid.credential({
            username: params.master_credentials.username,
            password: params.master_credentials.password,
            name: 'Env: '+env.uuid
          }).then((key) => {
            resolve({
              env: env,
              credentials: {
                SENDGRID_API_KEY: key.api_key
              }
            });
          }).catch(reject);
        });
      });

      return Promise.all(acquireKeys);
    }).then((credentials) => {    
      var actions = credentials.map((credential) => {
        return credentialsApi.create({
          session_token: params.session_token,
          type: 'env',
          owner: credential.env.uuid,
          name: 'SENDGRID_API_KEY',
          value: credential.credentials.SENDGRID_API_KEY
        });
      });

      return Promise.all(actions).then(resolve);
    }).catch(reject);
  });
};

credentials.list = function (params) {
  return new Promise((resolve, reject) => {
    vault.get().then((box) => {
      var opts = {
        session_token: box.get('session_token'),
        owner: params.owner,
        type: params.type 
      };

      return credentialsApi.list(opts).then(resolve);
    }).catch(reject);
  });
};
