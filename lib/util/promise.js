"use strict";

const promise = exports;

/**
 * Takes in a map of keys to promises and returns a promise. If all promises
 * are fulfilled then a map of the keys to results is returned. If any promise
 * is reject, then the returned promise is rejected with the passed reason.
 */
promise.map = function (obj) {
  return new Promise((resolve, reject) => {
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
