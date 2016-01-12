'use strict';

const fs = require('fs');
const path = require('path');

const resolver = exports;

resolver.expand = function expandSegments (base) {
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
 * Returns a list of paths that are direct parents of the base that contain
 * the given filename that the current process can read
 */
resolver.parents = function (base, name) {

  function check (filePath, mode) {
    return new Promise((resolve) => {
      fs.access(filePath, mode, function(err) {
        if (err) {
          return resolve(false);
        }

        resolve(filePath);
      });
    });
  }

  return new Promise((resolve, reject) => {

    var paths = resolver.expand(base);
    var checks = paths.map((dir) => {
      var filePath = path.join(dir, name);

      /*jshint bitwise: false*/
      return check(filePath, fs.F_OK | fs.R_OK);
    });

    Promise.all(checks).then((results) => {
      results = results.filter((file) => {
        return (typeof file === 'string');
      });

      resolve(results);
    }).catch(reject);
  });
};
