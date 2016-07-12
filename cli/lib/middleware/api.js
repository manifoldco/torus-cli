'use strict';

var api = require('../api');

module.exports = function () {
  return function (ctx) {
    ctx.api = api.build({ proxySocketUrl: ctx.config.proxySocketUrl });
    return true;
  };
};
