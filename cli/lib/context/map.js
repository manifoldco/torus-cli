'use strict';

var fs = require('fs');
var _ = require('lodash');
var path = require('path');
var Promise = require('es6-promise').Promise;

var resolver = require('../util/resolver');

var map = exports;

var MAP_FILE_NAME = '.arigato.json';
var MAP_DEFAULTS = {
  org: null,
  project: null
};

// Return map path for cwd
map.path = function () {
  return path.join(process.cwd(), MAP_FILE_NAME);
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
