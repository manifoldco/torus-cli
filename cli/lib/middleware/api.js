'use strict';

var api = require('../api');

module.exports = function () {
  return function (ctx) {
    ctx.api = api.build({
      socketUrl: ctx.config.socketUrl,
      registryUrl: ctx.prefs.values.core.registry_uri || ctx.config.registryUrl
    });
    return true;
  };
};
