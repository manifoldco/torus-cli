'use strict';

var lock = exports;

var Promise = require('es6-promise').Promise;
var lockfile = require('lockfile');

function getLockPath(file) {
  return file + '.lock';
}

lock.acquire = function (file) {
  return new Promise(function (resolve, reject) {
    var lockPath = getLockPath(file);

    /* eslint-disable consistent-return */
    lockfile.lock(lockPath, function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};

lock.release = function (file) {
  return new Promise(function (resolve, reject) {
    var lockPath = getLockPath(file);

    /* eslint-disable consistent-return */
    lockfile.unlock(lockPath, function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};
