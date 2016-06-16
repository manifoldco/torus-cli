'use strict';

var validate = require('../validate');

var harvest = module.exports;

var createValidator = validate.build({
  project: validate.slug,
  environment: validate.slug,
  service: validate.slug,
  org: validate.slug,
  name: validate.credName,
  instance: validate.slug
});

var getValidator = validate.build({
  org: validate.slug,
  project: validate.slug,
  environment: validate.slug,
  service: validate.slug,
  instance: validate.slug
});

/**
 * Harvests parameters from ctx and constructs an object
 * for usage with credentials.create.
 *
 * @param {Context} ctx cli context object
 * @return {Promise}
 */
harvest.create = function (ctx) {
  var name = ctx.params[0];
  var parts = name.split('/').filter(function (part) {
    return part.length !== 0;
  });

  // A CPathExp with the name is 7 distinct parts, without the name its 6.
  if (parts.length === 7) {
    // Add in the project component to the path.
    return {
      name: parts[6],
      path: '/' + parts.slice(0, 6).join('/')
    };
  }

  var orgName = ctx.option('org').value;
  var projectName = ctx.option('project').value;
  var serviceName = ctx.option('service').value;
  var envName = ctx.option('environment').value;
  var instance = ctx.option('instance').value;

  if (!orgName) {
    throw new Error('You must provide a --org flag');
  }

  if (!projectName) {
    throw new Error('You must provide a --project flag');
  }

  if (!serviceName) {
    throw new Error('You must provide a --service flag');
  }

  if (!envName) {
    throw new Error('You must provide a --environment flag');
  }

  var errors = createValidator({
    org: orgName,
    name: name,
    project: projectName,
    service: serviceName,
    environment: envName,
    instance: instance
  });
  if (errors.length > 0) {
    throw errors[0];
  }

  return {
    org: orgName,
    name: name,
    project: projectName,
    service: serviceName,
    environment: envName,
    instance: instance
  };
};

/**
 * Harvests parameters from ctx and constructs an object for usage with
 * credentials.get.
 *
 * @param {Context} ctx
 * @return {Promise}
 */
harvest.get = function (ctx) {
  var orgName = ctx.option('org').value;
  var projectName = ctx.option('project').value;
  var serviceName = ctx.option('service').value;
  var envName = ctx.option('environment').value;
  var instance = ctx.option('instance').value;

  if (!orgName) {
    throw new Error('You must provide a --org flag');
  }

  if (!projectName) {
    throw new Error('You must provide a --project flag');
  }

  if (!serviceName) {
    throw new Error('You must provide a --service flag');
  }

  if (!envName) {
    throw new Error('You must provide a --environment flag');
  }

  var params = {
    org: orgName,
    project: projectName,
    service: serviceName,
    environment: envName,
    instance: instance
  };

  var errors = getValidator(params);
  if (errors.length > 0) {
    throw errors[0];
  }

  return params;
};
