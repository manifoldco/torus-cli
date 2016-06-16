'use strict';


var harvest = module.exports;

var validate = require('../validate');
var Target = require('../context/target');

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
  if (!(ctx.target instanceof Target)) {
    throw new TypeError('Target must exist on the Context object');
  }

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

  ctx.target.flags({
    org: ctx.option('org').value,
    project: ctx.option('project').value,
    service: ctx.option('service').value,
    environment: ctx.option('environment').value
  });

  var data = {
    org: ctx.target.org,
    project: ctx.target.project,
    environment: ctx.target.environment,
    service: ctx.target.service,
    instance: ctx.option('instance').value,
    name: name
  };

  if (!data.org) {
    throw new Error('You must provide a --org flag');
  }

  if (!data.project) {
    throw new Error('You must provide a --project flag');
  }

  if (!data.service) {
    throw new Error('You must provide a --service flag');
  }

  if (!data.environment) {
    throw new Error('You must provide a --environment flag');
  }

  var errors = createValidator(data);
  if (errors.length > 0) {
    throw errors[0];
  }

  return data;
};

/**
 * Harvests parameters from ctx and constructs an object for usage with
 * credentials.get.
 *
 * @param {Context} ctx
 * @return {Promise}
 */
harvest.get = function (ctx) {
  ctx.target.flags({
    org: ctx.option('org').value,
    project: ctx.option('project').value,
    service: ctx.option('service').value,
    environment: ctx.option('environment').value
  });

  var data = {
    org: ctx.target.org,
    project: ctx.target.project,
    environment: ctx.target.environment,
    service: ctx.target.service,
    instance: ctx.option('instance').value
  };

  if (!data.org) {
    throw new Error('You must provide a --org flag');
  }

  if (!data.project) {
    throw new Error('You must provide a --project flag');
  }

  if (!data.service) {
    throw new Error('You must provide a --service flag');
  }

  if (!data.environment) {
    throw new Error('You must provide a --environment flag');
  }

  var errors = getValidator(data);
  if (errors.length > 0) {
    throw errors[0];
  }

  return data;
};
