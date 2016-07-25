/* eslint-disable consistent-return */
'use strict';

var fs = require('fs');
var Promise = require('es6-promise').Promise;

var fsWrap = exports;

fsWrap.stat = function (targetPath) {
  return new Promise(function (resolve, reject) {
    fs.stat(targetPath, function (err, stat) {
      if (err) {
        return reject(err);
      }

      resolve(stat);
    });
  });
};

fsWrap.read = function (targetPath) {
  return new Promise(function (resolve, reject) {
    fs.readFile(targetPath, function (err, contents) {
      if (err) {
        return reject(err);
      }

      resolve(contents);
    });
  });
};

fsWrap.write = function (targetPath, contents) {
  return new Promise(function (resolve, reject) {
    fs.writeFile(targetPath, contents, function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};
