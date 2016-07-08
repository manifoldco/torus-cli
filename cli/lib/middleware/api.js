'use strict';

var api = require('../api');

module.exports = function () {
  return function (ctx) {
    ctx.api = api.build();

    if (ctx.session && ctx.session.token) {
      ctx.api.auth(ctx.session.token);
    }

    return true;
  };
};
