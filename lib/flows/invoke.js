'use strict';

const schema = require('../util/schema');

const invoke = exports;

/**
 * Runs a piece of code from a descriptor package. Eventually this will be
 * safe -- right now it's not :)  
 */
invoke.method = function (params) {
  return new Promise((resolve, reject) => {
    var fn = require(params.method);
    var p = fn.apply(null, params.args);

    if (!p.then) {
      return reject(new Error('Did not return a promise: '+params.method));
    }

    p.then((results) => {
      return schema.validate(params.output, results)
        .then(resolve).catch(reject); 
    }).catch(reject);
  });
};
