'use strict';

var _ = require('lodash');
var api = require('../api');

module.exports = function () {
  return function (ctx) {
    ctx.api = api.build({
      socketUrl: ctx.config.socketUrl,
      registryUrl: _.get(ctx.prefs.values, 'core.registry_uri') || ctx.config.registryUrl
    });
    return true;
  };
};
