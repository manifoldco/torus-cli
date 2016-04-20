'use strict';

var Promise = require('es6-promise').Promise;

module.exports = function (list) {
  function iter (list, results) {
    var item = list.shift();
    if (!item) {
      return Promise.resolve(results);
    }

    return item().then(function (result) {
      results.push(result);
      return iter(list, results);
    });
  }

  return iter(list, []);
};
