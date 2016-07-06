'use strict';

var fs = require('fs');
var path = require('path');
var Promise = require('es6-promise').Promise;

var resolver = exports;

resolver.expand = function expandSegments(base) {
  base = path.normalize(base);
  var paths = [base];

  var chunk;
  var current = base;

  // a expand that works by sequencing on '/'
  for (var nextPlace = current.lastIndexOf('/');
       nextPlace > -1; nextPlace = current.lastIndexOf('/')) {
    chunk = current.slice(0, nextPlace); // get the new chunk
    paths.push(path.normalize(chunk)); // normalize it
    current = chunk; // widdle current down tot he remaining chunk
  }

  return paths;
};

/**
 * Returns a list of paths that are direct parents of the base that varain
 * the given filename that the current process can read
 */
resolver.parents = function (base, name) {
  function check(filePath, mode) {
    return new Promise(function (resolve) {
      return fs.access(filePath, mode, function (err) {
        if (err) {
          return resolve(false);
        }

        return resolve(filePath);
      });
    });
  }

  return new Promise(function (resolve, reject) {
    var paths = resolver.expand(base);
    var checks = paths.map(function (dir) {
      var filePath = path.join(dir, name);

      /* jshint bitwise: false*/
      return check(filePath, fs.F_OK | fs.R_OK);
    });

    return Promise.all(checks).then(function (results) {
      results = results.filter(function (file) {
        return (typeof file === 'string');
      });

      return resolve(results);
    }).catch(reject);
  });
};
