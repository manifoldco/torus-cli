'use strict';

const errors = require('./errors');
const vault = require('./util/vault');
const Arigato = require('./descriptors/arigato').Arigato;
const credentials = require('./credentials');
const servicesApi = require('./api/services');
const appsApi = require('./api/apps');

const services = exports;

services.add = function (params) {
  return new Promise((resolve, reject) => {
    return vault.get().then((box) => {
      return Arigato.find(process.cwd()).then((arigato) => {
        if (!arigato) {
          var msg ='An arigato.yml file does not exist, please init'+
            'in your project root';
          return reject(new errors.ArigatoConfigError(msg));
        }

        var exists = arigato.get('services').some((service) => {
          return service.toLowerCase() === params.name;
        });

        if (exists) {
          return reject(new errors.ArigatoConfigError(
            'That service already is a part of your project'
          ));
        }

        var appOpts = {
          session_token: box.get('session_token'),
          app_id: arigato.get('app')
        };
        return appsApi.get(appOpts).then((app) => {

          var opts = {
            session_token: box.get('session_token'),
            app: app,
            name: params.name, // TODO: Come back and figure htis out :)
            service: params.name
          };

          return servicesApi.add(opts).then((service) => {

            var serviceList = arigato.get('services') || [];
            serviceList.push(service.slug);

            return arigato.set('services', serviceList).then(() => {
              return arigato.write().then(() => {
                return credentials.initialize({
                  session_token: box.get('session_token'),
                  app: app,
                  service: service,
                  deployment: params.deployment,
                  descriptor: params.descriptor,
                  provided: params.provided
                }).then(() => {
                  resolve(service);
                });
              });
            });
          });
        });
      });
    }).catch(reject);
  });
};

services.get = function () {
  return new Promise((resolve, reject) => {
    return Promise.all([
      vault.get(),
      Arigato.find(process.cwd())
    ]).then((results) => {
      var box = results[0];
      var arigato = results[1];

      return servicesApi.get({
        session_token: box.get('session_token'),
        app: { uuid: arigato.get('app') }
      }).then(resolve);
    }).catch(reject);
  });
};

services.supported = function () {
  return servicesApi.supported();
};
