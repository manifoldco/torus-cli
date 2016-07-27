'use strict';

var validate = require('../validate');

var validator = validate.build({
  actions: validate.actionShorthand,
  path: validate.expResourcePath,
  team: validate.slug
});

/**
 * Harvests parameters from ctx and constructs an object for usage with
 * access methods.
 *
 * @param {Context} ctx
 * @return {Promise}
 */
function harvest(ctx) {
  var data = {
    actions: ctx.params[0],
    path: ctx.params[1],
    team: ctx.params[2]
  };

  if (!data.actions) {
    throw new Error('You must provide a [actions] arg');
  }

  if (!data.path) {
    throw new Error('You must provide a [path] arg');
  }

  if (!data.team) {
    throw new Error('You must provide a [team] arg');
  }

  var errors = validator(data);
  if (errors.length > 0) {
    throw errors[0];
  }

  return data;
}

module.exports = harvest;
