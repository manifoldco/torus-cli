'use strict';

var Promise = require('es6-promise').Promise;

var Config = require('../config');
var targetMap = require('../context/map');
var Target = require('../context/target');

module.exports = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      if (!(ctx.config instanceof Config)) {
        throw new TypeError('Config object must exist on Context object');
      }

      targetMap.derive(ctx.config, process.cwd()).then(function (matches) {
        if (matches.length > 0) {
          ctx.target = matches[0];
          return resolve();
        }

        ctx.target = new Target(process.cwd(), {
          org: null,
          project: null,
          service: null
        });

        return resolve();
      }).catch(reject);
    });
  };
};
