'use strict';

var fs = require('fs');

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var lock = require('../util/lock');
var Target = require('./target');
var Config = require('../config');

var map = exports;

// XXX: This should be 0600 not 0700
var MAP_PERM_STRING = '0700';
var MAP_FILE_PERM = 448;

/**
 * Links a given directory to a target. Holds a lock on the context map file so
 * no one else can write.
 *
 * @param {Config} config
 * @param {Target} target
 */
map.link = function (config, target) {
  if (!(config instanceof Config)) {
    throw new TypeError('Must provide a Config object');
  }
  if (!(target instanceof Target)) {
    throw new TypeError('Must provide a Target object');
  }

  return new Promise(function (resolve, reject) {
    lock.wrap(config.mapPath, function () {
      return map.get(config).then(function (targetMap) {
        var matches = map._getMatches(targetMap, target.path);

        // Check if the org/project differs for anything in the path
        var allMatch = matches.every(function (match) {
          // If the path is the *exact* same then it doesn't matter; we're
          // going to be overwriting this anyways.
          if (match.path === target.path) {
            return true;
          }

          return (match.org === target.org && match.project === target.project);
        });

        if (!allMatch) {
          throw new Error('sub-directories cannot link to different ' +
                          'org/projects than their parents');
        }

        targetMap[target.path] = target.context();
        return map._writeFile(config.mapPath, targetMap).then(function () {
          resolve(map._getMatches(targetMap, target.path));
        });
      });
    }).catch(reject);
  });
};

/**
 * Unlinks a given target. Holds a lock on the context map file so no one else
 * can write.
 *
 * @param {Config} config
 * @param {Target} target
 */
map.unlink = function (config, target) {
  if (!(config instanceof Config)) {
    throw new TypeError('Must provide a Config object');
  }

  if (!(target instanceof Target)) {
    throw new TypeError('Must provide a Target object');
  }

  return new Promise(function (resolve, reject) {
    lock.wrap(config.mapPath, function () {
      return map.get(config).then(function (targetMap) {
        if (!targetMap[target.path]) {
          throw new Error('Target path is not linked in ' + config.mapPath);
        }

        delete targetMap[target.path];
        return map._writeFile(config.mapPath, targetMap).then(function () {
          resolve(map._getMatches(targetMap, target.path));
        });
      });
    }).catch(reject);
  });
};

/**
 * Returns a sorted array of matching Target objects. The first object is the
 * one target that is most relevant.
 *
 * @param {Config} config
 * @param {String} dir
 * @returns Promise resolves with an array of Target objects
 */
map.derive = function (config, dir) {
  if (!(config instanceof Config)) {
    throw new TypeError('Must provide a Config object');
  }

  return new Promise(function (resolve, reject) {
    map.get(config).then(function (targetMap) {
      resolve(map._getMatches(targetMap, dir));
    }).catch(reject);
  });
};

/**
 * Returns the raw contents of the context map file.
 */
map.get = function (config) {
  if (!(config instanceof Config)) {
    throw new TypeError('Must provide a config object');
  }

  return new Promise(function (resolve, reject) {
    /* eslint-disable consistent-return, no-shadow */
    fs.stat(config.mapPath, function (err, stat) {
      if (err && err.code === 'ENOENT') {
        return resolve({});
      }
      if (err) {
        return reject(err);
      }

      if (!stat.isFile()) {
        return reject(new Error(
          'Context map file must be a file: ' + config.mapPath));
      }

      var fileMode = '0' + (stat.mode & parseInt('777', 8)).toString(8);
      if (fileMode !== MAP_PERM_STRING) {
        return reject(new Error('Context map file permission error: ' +
          config.mapPath + ' ' + fileMode + ' not ' + MAP_PERM_STRING));
      }

      fs.readFile(config.mapPath, 'utf-8', function (err, contents) {
        if (err) {
          return reject(err);
        }

        var data;
        try {
          data = JSON.parse(contents);
        } catch (err) {
          return reject(err);
        }

        return resolve(data);
      });
    });
  });
};

map._getMatches = function (targetMap, dir) {
  var pathMatchers = {};
  Object.keys(targetMap).forEach(function (path) {
    pathMatchers[path] = new RegExp('^' + path);
  });

  var matches = {};
  Object.keys(pathMatchers).forEach(function (path) {
    if (!pathMatchers[path]) {
      throw new Error('unknown matcher for path: ' + path);
    }

    if (!pathMatchers[path].test(dir)) {
      return;
    }

    var distance = dir.length - path.length;
    if (distance === -1) {
      throw new Error('Should not happen, cwd path shorter than map: ' + path);
    }

    matches[path] = {
      path: path,
      distance: distance
    };
  });

  var sorted = _.sortBy(matches, 'distance').map(function (element) {
    return new Target(element.path, targetMap[element.path]);
  });

  return sorted;
};

map._writeFile = function (mapPath, targetMap) {
  return new Promise(function (resolve, reject) {
    var contents = JSON.stringify(targetMap);
    var opts = {
      encoding: 'utf-8',
      mode: MAP_FILE_PERM
    };

    fs.writeFile(mapPath, contents, opts, function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};
