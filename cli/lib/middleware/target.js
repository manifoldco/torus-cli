'use strict';

var Promise = require('es6-promise').Promise;

var targetMap = require('../context/map');
var Target = require('../context/target');

module.exports = function () {
  return function (ctx) {
    return new Promise(function (resolve, reject) {
      targetMap.get().then(function (result) {
        ctx.target = new Target(result);

        return resolve();
      }).catch(reject);
    });
  };
};
