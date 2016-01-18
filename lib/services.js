'use strict';

const errors = require('./errors');
const validation = require('./util/validation');
const vault = require('./util/vault');
const Arigato = require('./arigato').Arigato;
const credentials = require('./credentials');
const servicesApi = require('./api/services');

const services = exports;

services.add = function (params) {
  return new Promise((resolve, reject) => {
    var validMsg = validation.serviceName.validate(params.serviceName);
    if (typeof validMsg === 'string') {
      return reject(new errors.ValidationError(validMsg));
    }

    return vault.get().then((box) => {
      return Arigato.find(process.cwd()).then((arigato) => {
        if (!arigato) {
          var msg ='An arigato.yml file does not exist, please init'+
            'in your project root';
          return reject(new errors.ArigatoConfigError(msg));
        }

        var serviceName = params.serviceName.toLowerCase();
        var exists = arigato.get('services').some((service) => {
          return service.toLowerCase() === serviceName;
        });

        if (exists) {
          return reject(new errors.ArigatoConfigError(
            'That service already is a part of your project'
          ));
        }

        var opts = {
          session_token: box.get('session_token'),
          app_id: arigato.get('app'), // TODO: We need slug based lookup
          name: serviceName, // TODO: Come back and figure htis out :)
          service: serviceName
        };

        return servicesApi.add(opts).then((service) => {

          var serviceList = arigato.get('services');
          serviceList.push(service.slug);

          return arigato.set('services', serviceList).then(() => {
            return arigato.write().then(() => {
              return credentials.initialize({
                session_token: box.get('session_token'),
                app_id: arigato.get('app'),
                service: service,
                master_credentials: params.credentials
              }).then(() => {
                resolve(service);
              });
            });
          });
        });
      });
    }).catch(reject);
  });
};

services.supported = function () {
  return servicesApi.supported();
};
