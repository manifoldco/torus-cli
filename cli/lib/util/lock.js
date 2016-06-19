'use strict';

var lock = exports;

var Promise = require('es6-promise').Promise;
var lockfile = require('lockfile');

function getLockPath(file) {
  return file + '.lock';
}

lock.wrap = function (file, fn) {
  return new Promise(function (resolve, reject) {
    var locked;
    return lock.acquire(file).then(function () {
      locked = true;

      var args;
      return fn().then(function () {
        args = arguments;
        return lock.release(file);
      }).then(function () {
        resolve.apply(resolve, args);
      });
    }).catch(function (err) {
      if (!locked) {
        return reject(err);
      }

      return lock.release(file).then(function () {
        return reject(err);
      }).catch(function (releaseErr) {
        return reject(new Error(
          'Could not release file: ' + file + ' err: ' + releaseErr.message));
      });
    });
  });
};

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
