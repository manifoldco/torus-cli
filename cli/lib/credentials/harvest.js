'use strict';

var harvest = module.exports;

var _ = require('lodash');
var errors = require('common/errors');

var validate = require('../validate');
var Target = require('../context/target');

var createValidator = validate.build({
  project: validate.slug,
  environment: validate.orExpression,
  service: validate.orExpression,
  org: validate.slug,
  name: validate.credName,
  identity: validate.orExpression,
  instance: validate.orExpression
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
  if (typeof name !== 'string') {
    throw new TypeError('Expected the first paramter to be a string');
  }

  name = name.toLowerCase();
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
  } else if (parts.length > 1) {
    throw new Error('Invalid path: must be 7 segments');
  }

  ctx.target.flags({
    org: ctx.option('org').value,
    project: ctx.option('project').value,
    environment: ctx.option('environment').value,
    service: ctx.option('service').value
  });

  var builtFlags = harvest._createFlags({
    service: ctx.target.service,
    environment: ctx.target.environment
  });

  var data = {
    org: ctx.target.org,
    project: ctx.target.project,
    service: builtFlags.service,
    environment: builtFlags.environment,
    instance: ctx.option('instance').value.toString(),
    identity: ctx.option('user').value || '*',
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

  var errs = createValidator(data);
  if (errs.length > 0) {
    throw errs[0];
  }

  return data;
};

/**
 * For creation only, we pull the service and environment flags and then create
 * OR expressions if they've been specified more than once.
 */
harvest._createFlags = function (data) {
  var keys = ['service', 'environment'];
  var flags = {};
  keys.forEach(function (key) {
    var item = data[key];
    if (!item) {
      throw new errors.Validation('You must provide a --' + key + ' flag');
    }

    item = Array.isArray(item) ? item : [item];

    // Validate the flags
    item.forEach(function (v) {
      var err = validate.slugOrWildcard(v);
      if (_.isString(err)) {
        throw new errors.Validation('invalid flag ' + key + ': ' + err);
      }

      if (item.length > 1 && v === '*') {
        throw new errors.Validation(
          'invalid flag ' + key + ': cannot contain * wild multiple invocations');
      }
    });

    // Turn into an OR if more than one is provided
    if (item.length === 1) {
      flags[key] = item[0];
      return;
    }

    flags[key] = '[' + item.join('|') + ']';
  });

  return flags;
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
    instance: ctx.option('instance').value.toString()
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

  var errs = getValidator(data);
  if (errs.length > 0) {
    throw errs[0];
  }

  return data;
};
