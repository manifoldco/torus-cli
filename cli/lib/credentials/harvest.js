'use strict';

var validate = require('../validate');
var validator = validate.build({
  environment: validate.slug,
  service: validate.slug,
  name: validate.credName,
  instance: validate.slug
});

/**
 * Harvests parameters from ctx and constructs an object
 * for usage with credentials.create.
 *
 * @param {Context} ctx cli context object
 * @return {Promise}
 */
module.exports = function (ctx) {
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

  var serviceName = ctx.option('service').value;
  var envName = ctx.option('environment').value;
  var instance = ctx.option('instance').value;

  if (!serviceName) {
    throw new Error('You must provide a --service flag');
  }

  if (!envName) {
    throw new Error('You must provide a --environment flag');
  }

  var errors = validator({
    name: name,
    service: serviceName,
    environment: envName,
    instance: instance
  });
  if (errors.length > 0) {
    throw errors[0];
  }

  return {
    name: name,
    service: serviceName,
    environment: envName,
    instance: instance
  };
};
