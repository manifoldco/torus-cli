"use strict";

const _ = require('lodash');
const promise = exports;

promise.map = function (obj) {
  return new Promise((resolve, reject) => {
    
    if (!_.isPlainObject(obj)) {
      throw new TypeError('Must be a plain object');
    }

    var promises = [];

    Object.keys(obj).forEach((key) => {
      var value = obj[key];
      promises.push(new Promise((resolve, reject) => {
        value.then((result) => {
          resolve({
            key: key,
            value: result
          }); 
        }).catch(reject); 
      }));
    });

    Promise.all(promises).then((results) => {
      var map = {};
      results.forEach((result) => {
        map[result.key] = result.value;
      });

      resolve(map);
    }).catch(reject);
  });
};

promise.series = function (list) {
  function iter (list, results) {
    var item = list.shift();
    if (!item) {
      return Promise.resolve(results);
    }

    return item().then((result) => {
      results.push(result);
      return iter(list, results);
    });
  }

  return iter(list, []);
};

promise.seriesMap = function (obj) {
  function iter (keys, results) {
    var key = keys.shift();
    if (!key) {
      return Promise.resolve(results);
    }

    return obj[key]().then((result) => {
      results[key] = result;
      return iter(keys, results);
    });
  }

  return iter(Object.keys(obj), {});
};
