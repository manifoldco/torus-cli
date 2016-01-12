'use strict';

const validation = require('./util/validation');
const errors = require('./errors');
const Arigato = require('./arigato').Arigato;
const vault = require('./util/vault');
const appsApi = require('./api/apps');
const envsApi = require('./api/envs');

const apps = exports;

apps.init = function (appName) {
  return new Promise((resolve, reject) => {
    var validMsg = validation.appName.validate(appName);
    if (typeof validMsg === 'string') {
      return reject(new errors.ValidationError(validMsg));
    }

    Arigato.find(process.cwd()).then((arigato) => {
      if (arigato) {
        var msg = 'An arigato.yml file already exists: '+arigato.path;
        return reject(new errors.ArigatoConfigError(msg, 'file_exists_error'));
      }

      return vault.get();
    }).then((box) => {
      var opts = {
        session_token: box.get('session_token'),
        name: appName
      };

      return appsApi.create(opts).then((app) => {
        return Arigato.create(process.cwd(), {
          app: app,
          user_id: box.get('uuid')
        }).then((arigato) => {
          /*jshint unused:false*/
          var opts = {
            session_token: box.get('session_token'),
            app_id: app.uuid,
            name: box.get('uuid') // name the env after the user for now
          };

          return envsApi.create(opts).then((env) => {
            resolve({
              app: app,
              env: env
            }); 
          });
        });
      });
    }).catch(reject);
  });
};
