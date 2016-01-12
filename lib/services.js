'use strict';

const servicesApi = require('./api/services');
const errors = require('./errors');
const validation = require('./util/validation');
const vault = require('./util/vault');
const Arigato = require('./arigato').Arigato;

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
        var exists = arigato.services.some((service) => {
          return service.toLowerCase() === serviceName; 
        });

        if (exists) {
          return reject(new errors.ArigatoConfigError(
            'That service already is a part of your project'
          ));
        }

        var opts = {
          session_token: box.get('session_token'),
          app_id: arigato.app, // TODO: We need slug based lookup
          name: serviceName, // TODO: Come back and figure htis out :)
          service: serviceName
        };

        return servicesApi.add(opts).then((service) => {
          arigato.addService(service.slug);
          return arigato.write().then(() => {
            /// arlgiht now create credentials for each environment here by
            // talking to the sendgrid api wrapper that jeff made.
            //
            // and upload them + the master creds to the api.
            resolve(service);
          });
        });
      });
    }).catch(reject);
  });  
};

services.supported = function () {
  return servicesApi.supported();
};
