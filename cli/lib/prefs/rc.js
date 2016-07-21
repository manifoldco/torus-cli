'use strict';

var _ = require('lodash');
var path = require('path');
var ini = require('ini');
var fs = require('fs');
var lock = require('../util/lock');
var Prefs = require('../prefs');
var Promise = require('es6-promise').Promise;

// XXX: This should be 0600 not 0700
var RC_PERM_STRING = '0644';

var rc = module.exports = {};

rc.stat = function (rcPath) {
  return new Promise(function (resolve, reject) {
    fs.stat(rcPath, function (err, stat) {
      if (err && err.code === 'ENOENT') {
        return resolve(false);
      }

      if (err) {
        return reject(err);
      }

      if (!stat.isFile()) {
        return reject(new Error('.arigatorc must be a file: ' + rcPath));
      }

      var fileMode = '0' + (stat.mode & parseInt('777', 8)).toString(8);

      if (fileMode !== RC_PERM_STRING) {
        var msg = 'rc file permission error: ' + rcPath + ' ' + fileMode + ' not ' + RC_PERM_STRING;
        return reject(new Error(msg));
      }

      return resolve(true);
    });
  });
};

rc.write = function (prefs) {
  if (!(prefs instanceof Prefs)) {
    return Promise.reject(new Error('prefs must be an instance of Prefs'));
  }

  return lock.wrap(prefs.path, function () {
    return rc.stat(prefs.path).then(function () {
      return rc._write(prefs.path, prefs.values);
    });
  });
};

rc._write = function (rcPath, contents) {
  return new Promise(function (resolve, reject) {
    fs.writeFile(rcPath, ini.stringify(contents), function (err) {
      if (err) {
        return reject(err);
      }

      return resolve();
    });
  });
};

rc.read = function (rcPath) {
  if (!_.isString(rcPath)) {
    throw new TypeError('Must provide rcPath string');
  }

  if (!path.isAbsolute(rcPath)) {
    throw new Error('Must provide an absolute rc path');
  }

  return lock.wrap(rcPath, function () {
    return rc.stat(rcPath).then(function (exists) {
      if (!exists) {
        return {};
      }

      return rc._read(rcPath);
    });
  });
};

rc._read = function (rcPath) {
  return new Promise(function (resolve, reject) {
    fs.readFile(rcPath, 'utf8', function (err, contents) {
      if (err) {
        return reject(err);
      }

      return resolve(ini.parse(contents));
    });
  });
};
