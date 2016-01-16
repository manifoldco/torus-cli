'use strict';

const vault = require('./util/vault');
const envsApi = require('./api/envs');

const envs = exports;

envs.list = function (params) {
  return new Promise((resolve, reject) => {
    vault.get().then((box) => {

      var opts = {
        session_token: box.get('session_token'),
        app_id: params.app_id
      };

      return envsApi.list(opts).then(resolve);
    }).catch(reject);
  });
};
