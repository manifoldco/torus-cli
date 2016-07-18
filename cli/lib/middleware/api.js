'use strict';

var api = require('../api');

module.exports = function () {
  return function (ctx) {
    ctx.api = api.build({ socketUrl: ctx.config.socketUrl });
    return true;
  };
};
