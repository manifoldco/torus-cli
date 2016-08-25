'use strict';

var _ = require('lodash');
var api = require('../api');

module.exports.preHook = function () {
  return function (ctx) {
    var config = ctx.config;

    ctx.api = api.build({
      registryUrl: _.get(ctx, 'ctx.prefs.values.core.registry_uri', config.registryUrl),
      socketUrl: config.socketUrl,
      socketPath: config.socketPath
    });

    return true;
  };
};
