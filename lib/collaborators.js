'use strict';

const vault = require('./util/vault');
const Arigato = require('./arigato').Arigato;
const envsApi = require('./api/envs');
const collaboratorsApi = require('./api/collaborators');
const errors = require('./errors');
const validation = require('./util/validation');

const collaborators = exports;

collaborators.add = function (params) {
  return new Promise((resolve, reject) => {

    var validMsg = validation.email.validate(params.email);
    if (typeof validMsg === 'string') {
      return reject(new errors.ValidationError(validMsg));
    }

    return vault.get().then((box) => {
      return Arigato.find(process.cwd()).then((arigato) => {
        if (!arigato) {
          return reject(new errors.ArigatoConfigError(
            'Cannot locate arigato.yml config'
          ));
        }

        var opts = {
          email: params.email,
          app_id: arigato.app,
          session_token: box.get('session_token')
        };

        return collaboratorsApi.add(opts);
      }).then((results) => {

        var opts = {
          session_token: box.get('session_token'),
          name: results.user.uuid,
          app_id: results.app.uuid
        };

        return envsApi.create(opts).then((env) => {
          return resolve({
            env: env,
            user: results.user,
            app: results.app
          });
        });
      }); 
    }).catch(reject);
  });
};
