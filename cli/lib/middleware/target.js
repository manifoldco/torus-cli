'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var targetMap = require('../context/map');
var Target = require('../context/target');

module.exports = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      // Retrieve the context map
      targetMap.get().then(function (result) {
        // Detect if prefs has disabled context
        result.disabled = !_.get(ctx, 'prefs.core.context', true);

        // Nullify context values if not enabled
        if (result.disabled) {
          result.context = null;
        }

        ctx.target = new Target(result);

        // Context is disabled, proceed no further
        if (ctx.target.disabled()) {
          return resolve();
        }

        // Apply default values from prefs
        ctx.target.defaults(_.get(ctx, 'prefs.default', {}));

        // Apply context related flags to the target
        var service = ctx.option('service') || {};
        var environment = ctx.option('environment') || {};
        ctx.target.flags({
          service: service.value,
          environment: environment.value
        });

        // Look up user's default environment
        if (ctx.session && !ctx.target.environment) {
          return ctx.api.users.self().then(function (res) {
            var user = res && res[0];
            if (!user) {
              return reject(new Error('Could not find the user'));
            }

            ctx.target.flags({
              environment: 'dev-' + user.body.username
            });

            return resolve();
          });
        }

        return resolve();
      }).catch(reject);
    });
  };
};
