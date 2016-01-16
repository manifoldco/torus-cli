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
        name: 'sendgrid_username',
        value: params.master_credentials.username
      }),
      credentialsApi.create({
        session_token: params.session_token,
        type: 'app',
        owner: params.app_id,
        name: 'sendgrid_password',
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

function pickCredential (credName, credentials) {
  var cred = credentials.find((cred) => {
    return (cred.name === credName);
  });

  return (cred) ? cred.value : null;
}

credentials.create = function (params) {
  return new Promise((resolve, reject) => {
    vault.get().then((box) => {

      var opts = {
        session_token: box.get('session_token'),
        owner: params.app.uuid,
        type: 'app'
      };

      return credentialsApi.list(opts).then((creds) => {
        if (creds.length < 1) {
          return reject(new Error('Master creds dont exist for this service'));
        }

        var opts = {
          username: pickCredential('sendgrid_username', creds),
          password: pickCredential('sendgrid_password', creds),
          name: 'env:'+params.env.slug
        };

        return sendgrid.credential(opts);
      }).then((key) => {
        return credentialsApi.create({
          session_token: box.get('session_token'),
          type: 'env',
          owner: params.env.uuid,
          name: 'SENDGRID_API_KEY',
          value: key.api_key
        }).then((credential) => {
          return resolve([credential]);
        });
      });
    }).catch(reject);
  });
};
