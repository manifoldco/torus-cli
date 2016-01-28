'use strict';

var credentialsApi = require('./api/credentials');
var envsApi = require('./api/envs');
var vault = require('./util/vault');
var credentialFlow = require('./flows/credentials');
var questions = require('./flows/questions');
var ServiceRegistry = require('./descriptors/registry').Registry;

var credentials = exports;

credentials.initialize = function (params) {
  return new Promise((resolve, reject) => {

    // Build up the list of creds that need to be stored under the app
    // TODO: This should probably be pushed into the same code as
    // `lib/flow/credentials` so it's dry'd up
    var actions = [];
    Object.keys(params.provided).forEach((name) => {
      var credential = params.provided[name];
      Object.keys(credential.answers).forEach((componentName) => {
        var serviceName = params.service.name;

        actions.push(credentialsApi.create({
          session_token: params.session_token,
          type: 'app',
          owner: params.app.uuid,

          // component name, such as username or password -- meta helps
          // identify the owner
          name: componentName,
          value: credential.answers[componentName],
          meta: {
            service: serviceName,
            type: credential.type,

            // id is the "name" given tp the credential in the descriptor since
            // it's provided by the user
            id: credential.name
          }
        }));
      });
    });

    return Promise.all(actions).then(() => {
      var getEnvs = {
        session_token: params.session_token,
        app_id: params.app.uuid
      };

     return envsApi.list(getEnvs);
    }).then((envs) => {
      var credsToCreate = credentialFlow.determine(params.descriptor, envs);

      // XXX: This shouldnt' happen here.. a credential tree should be a first
      // class citizen
      var credTree = questions.toCredTree(params.descriptor, params.provided);
      return credentialFlow.acquire({
        list: credsToCreate,
        deployment: params.deployment,
        descriptor: params.descriptor,
        provided: credTree,
        app: params.app
      });
    }).then((credentials) => {

      var actions = [];
      credentials.forEach((cred) => {
        cred.components.forEach((component) => {
          actions.push(credentialsApi.create({
            session_token: params.session_token,
            type: 'env',
            owner: cred.env.uuid,
            name: component.name,
            value: component.value,
            meta: component.meta
          }));
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

// Unpacks an array of credentials into service, cred type, components
function unpackCredentials (provided, services) {

  var credTree = {};
  provided.forEach((component) => {
    var serviceName = component.meta.service;
    var service = services.find((service) => {
      return (service.name === component.meta.service);
    });

    if (!service) {
      throw new Error(
        'Could not find descriptor for service: '+serviceName);
    }

    var cred = service.descriptor.get('credentials').find((cred) => {
      return (component.meta.type === cred.name);
    });

    if (!cred) {
      throw new Error('Could not find matching cred type for: '+component.name);
    }

    if (!credTree[serviceName]) {
      credTree[serviceName] = {};
    }
    if (!credTree[serviceName][cred.meta.id]) {
      credTree[serviceName][cred.meta.id] = {};
    }

    credTree[serviceName][cred.meta.id][component.name] = component.value;
  });

  return credTree;
}

credentials.create = function (params) {
  return new Promise((resolve, reject) => {

    var registry = new ServiceRegistry();
    var getDescriptors = params.services.map((service) => {
      return new Promise((resolve, reject) => {
        return registry.get(service.name).then((deployment) => {
          return deployment.descriptor().then((descriptor) => {
            resolve({
              service: service,
              deployment: deployment,
              descriptor: descriptor
            });
          });
        }).catch(reject);
      });
    });

    return Promise.all([
      vault.get(),
      Promise.all(getDescriptors)
    ]).then((results) => {
      var box = results[0];
      var services = results[1];

      var opts = {
        session_token: box.get('session_token'),
        owner: params.app.uuid,
        type: 'app'
      };
      return credentialsApi.list(opts).then((provided) => {

        var required = {};
        var credTree = unpackCredentials(provided);
        var serviceMap = {};
        services.forEach((service) => {
          serviceMap[service.name] = service;
          required[service.name] = credentialFlow.determine(
            service.descriptor, [params.env]);
        });

        var toCreate = [];
        Object.keys(required).forEach((serviceName) => {
          toCreate.push(new Promise((resolve, reject) => {
            return credentialFlow.acquire({
              list: required[serviceName],
              deployment: params.deployment,
              descriptor: params.descriptor,
              provided: credTree,
              app: params.app
            }).then((results) => {
              resolve({
                deployment: serviceMap[serviceName].deployment,
                descriptor: serviceMap[serviceName].descriptor,
                acquired: results
              });
            }).catch(reject);
          }));
        });

        return Promise.all(toCreate);
      }).then((results) => {
        var actions = [];

        results.forEach((result) => {
          result.acquired.forEach((component) => {
            actions.push(credentialsApi.create({
              session_token: box.get('session_token'),
              type: 'env',
              owner: params.env.uuid,
              name: component.name,
              value: component.value,
              meta: component.meta
            }));
          });
        });

        return Promise.all(actions).then(resolve);
      });
    }).catch(reject);
  });
};
