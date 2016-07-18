'use strict';

/**
 * Attempt logout
 *
 * @param {object} ctx - Prompt context
 */
module.exports = function (ctx) {
  return ctx.api.logout.post();
};
