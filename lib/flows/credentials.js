'use strict';

const invoke = require('./invoke');

const credentials = exports;

credentials.acquire = function (params) {
  var toProcess = params.list.map((toCreate) => {
    switch (toCreate.type) {
      case 'copy':
        return credentials._copy(params, toCreate);

      case 'generated':
        return credentials._generate(params, toCreate);

      default:
        throw new Error('Unknown creation type: '+toCreate.type);
    }
  });

  return Promise.all(toProcess);
};

credentials._copy = function (params, toCreate) {
  return new Promise((resolve, reject) => {
    var providedType = params.descriptor.get(
      'default.provided')[toCreate.source];
    if (!providedType) {
      return reject(new Error(
        'Provided credential cannot be found: '+toCreate.source
      ));
    }

    var serviceName = params.descriptor.get('name');
    var provided = params.provided[serviceName];
    if (!provided) {
      return reject(new Error(
        'Service does not exist in cred tree: '+serviceName
      ));
    }

    var cred = provided[toCreate.source];
    if (!cred) {
      return reject(new Error(
        'Source does not exist in cred tree: '+toCreate.source
      ));
    }

    var components = [];
    Object.keys(cred).forEach((k) => {
      components.push({
        name: (serviceName+'_'+k).toUpperCase(),
        value: cred[k],
        meta: {
          service: serviceName,
          type: providedType.type,
          component: k,
          parent: toCreate.source
        }
      });
    });

    resolve({
      env: toCreate.env,
      app: params.app,
      components: components
    });
  });
};

credentials._generate = function (params, toCreate) {
  return new Promise((resolve, reject) => {
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
            component: k,
            parent: null
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
  var provided = descriptor.get('default.provided');

  var create = []; // list of creds to create

  envTypes.forEach((envType) => {
    // alright now look at sharing and the credential type, if env
    var credType = credTypes.find((cred) => {
      if (envType.credential) {
        return cred.name === envType.credential;
      }

      if (envType.source) {
        return cred.name === provided[envType.source].type;
      }

      return false;
    });

    if (!credType) {
      throw new Error('Cannot find credential for: '+envType.name);
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

  // shared & sourced from hierarchy (generated) -global value
  // shared & sourced from hierarchy (manual) - globl value
  // shared & no hierarchy (manual) - spans env type
  // shared & no hierarchy (generated) - spans env typ
  // shared & no hierachy (manual) - specific to env
  // shared & no hierarchy (generated) - specific to env

  // TODO: Come back and support the notion of "cred hierarchy".. you can
  // either inherit a shared cred or someone can come and input one. Right now
  // we only pull creds from the 'app".
  if (!(envType.source && credType.existence === 'singular' &&
        credType.creation === 'manual')) {
    throw new Error(
      'Cannot create shared cred thats not inherited: '+envType.name);
  }


  return {
    type: 'copy',
    source: envType.source,
    env: env
  };
};
