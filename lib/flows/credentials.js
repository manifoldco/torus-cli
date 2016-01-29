'use strict';

const invoke = require('./invoke');

const credentials = exports;

credentials.acquire = function (params) {
  var toProcess = params.list.map((toCreate) => {
    return new Promise((resolve, reject) => {
      // Don't support 'shared' credentials at the moment
      if (toCreate.type === 'shared') {
        throw new Error('Cannot create shared credentials');
      }

      var credential = params.descriptor.get('credentials').find((cred) => {
        return cred.name === toCreate.call.output;
      });

      if (!credential) {
        throw new Error(
          'Cannot invoke, no credential name: '+toCreate.call.output);
      }

      var opts = {
        method: toCreate.call.src,
        output: credential.schema,
        args: [params.deployment, params.descriptor, {
          app: params.app,
          env: toCreate.env,
          provided: params.provided[params.descriptor.get('name')]
        }]
      };

      return invoke.method(opts).then((results) => {
        var components = [];
        Object.keys(results).forEach((k) => {
          components.push({
            name: (params.descriptor.get('name')+'_'+k).toUpperCase(),
            value: results[k],
            meta: {
              service: params.descriptor.get('name'),
              type: credential.name,
              component: k
            }
          });
        });

        resolve({
          env: toCreate.env,
          app: params.app,
          components: components
        });
      }).catch(reject);
    });
  });

  return Promise.all(toProcess);
};

/**
 * Determines which credentials need to be generated and for which environment
 * instances.
 *
 * Currently assumes only development credentials are created and that all
 * environments are development environments.
 */
credentials.determine = function (descriptor, envs) {
  // look at descriptor and loop over each environment configuration
  // assume only development "type" today.
  var envTypes = descriptor.get('default.environments');
  var credTypes = descriptor.get('credentials');
  var create = []; // list of creds to create

  envTypes.forEach((envType) => {
    // alright now look at sharing and the credential type
    var credType = credTypes.find((cred) => {
      return cred.name === envType.credential;
    });

    // TODO: This shouldn't even be possible if we validated the descriptors
    // properly
    if (!credType) {
      throw new Error('Cannot find credential type: '+envType.credential);
    }

    create = create.concat(envs.map((env) => {
      // XXX Schema ensures it must either be unique or shared
      return credentials[envType.sharing](envType, credType, env);
    }));
  });

  return create;
};

credentials.unique = function (envType, credType, env) {

  // TODO: This should be done as a part of validation of a descriptor
  if (credType.existence !== 'multiple') {
    throw new Error(
      'Cannot create uniq cred, if multiple cannot exist:'+credType.name);
  }
  if (credType.creation !== 'generated') {
    throw new Error(
      'Cannot create uniq cred, if creation is not automated: '+credType.name);
  }

  var createFn = credType.methods.find((method) => {
    return method.type === 'create';
  });

  // TODO: This should also be a part of validation of a descriptor
  if (!createFn) {
    throw new Error(
      'Cannot create uniq cred, requires create method'
    );
  }

  return {
    type: 'generated',
    call: createFn,
    env: env
  };
};

credentials.shared = function (envType, credType, env) {
  /*jshint unused:false*/
  throw new Error('Not Implemented');
};
