'use strict';

var fs = require('fs');
var _ = require('lodash');
var path = require('path');
var Promise = require('es6-promise').Promise;

var lock = require('../util/lock');
var Target = require('./target');
var resolver = require('../util/resolver');

var map = exports;

var MAP_FILE_NAME = '.arigato.json';
var MAP_DEFAULTS = {
  org: null,
  project: null
};
var MAP_KEYS = _.keys(MAP_DEFAULTS);

// Return map path for cwd
map.path = function () {
  return path.join(process.cwd(), MAP_FILE_NAME);
};

/**
 * Links a given directory to a target. Holds a lock on the context map file so
 * no one else can write.
 *
 * @param {Target} target
 */
map.link = function (target) {
  if (!(target instanceof Target)) {
    throw new TypeError('Must provide a Target object');
  }

  return new Promise(function (resolve, reject) {
    var targetPath = target.path();
    lock.wrap(targetPath, function () {
      return map._writeFile(targetPath, _.pick(target.context(), MAP_KEYS))
        .then(resolve);
    }).catch(reject);
  });
};

/**
 * Unlinks a given target. Holds a lock on the context map file so no one else
 * can write.
 *
 * @param {Target} target
 */
map.unlink = function (target) {
  if (!(target instanceof Target)) {
    throw new TypeError('Must provide a Target object');
  }

  return new Promise(function (resolve, reject) {
    if (!target.exists()) {
      throw new Error('Target path is not linked at ' + target.path());
    }

    return map._rmFile(target.path())
      .then(resolve)
      .catch(reject);
  });
};

/**
 * Retrieves contents and path of the context map file
 */
map.get = function () {
  return new Promise(function (resolve, reject) {
    var mapPath = map.path();
    resolver.parents(process.cwd(), MAP_FILE_NAME).then(function (files) {
      if (files.length === 0) {
        return resolve({
          path: mapPath,
          context: null
        });
      }

      mapPath = files[0];
      return map._readFile(mapPath)
        .then(function (contents) {
          resolve({
            context: contents,
            path: files[0]
          });
        })
        .catch(reject);
    }).catch(reject);
  });
};

/**
 * Returns the raw contents of the context map file.
 */
map._readFile = function (mapPath) {
  return new Promise(function (resolve, reject) {
    /* eslint-disable consistent-return, no-shadow */
    fs.stat(mapPath, function (err, stat) {
      if (err && err.code === 'ENOENT') {
        return resolve(_.defaults({}, MAP_DEFAULTS));
      }
      if (err) {
        return reject(err);
      }

      if (!stat.isFile()) {
        return reject(new Error(
          'Context map file must be a file: ' + mapPath));
      }

      fs.readFile(mapPath, 'utf-8', function (err, contents) {
        if (err) {
          return reject(err);
        }

        var data;
        try {
          data = JSON.parse(contents);
        } catch (err) {
          return reject(err);
        }

        var requiredFields = Object.keys(MAP_DEFAULTS);
        var missingOrInvalid = _.filter(requiredFields, function (field) {
          return _.isUndefined(data[field]) || !_.isString(data[field]);
        });

        if (missingOrInvalid.length > 0) {
          throw new Error('Invalid ' + MAP_FILE_NAME);
        }

        return resolve(_.defaults(data, MAP_DEFAULTS));
      });
    });
  });
};

map._writeFile = function (mapPath, targetMap) {
  return new Promise(function (resolve, reject) {
    var contents = JSON.stringify(targetMap, null, 2);
    var opts = { encoding: 'utf-8' };

    fs.writeFile(mapPath, contents, opts, function (err) {
      if (err) {
        return reject(err);
      }

      resolve();
    });
  });
};

map._rmFile = function (mapPath) {
  return new Promise(function (resolve, reject) {
    fs.unlink(mapPath, function (err) {
      if (err) {
        return reject(err);
      }
      resolve();
    });
  });
};
