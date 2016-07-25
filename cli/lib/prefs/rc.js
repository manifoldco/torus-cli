'use strict';

var _ = require('lodash');
var path = require('path');
var ini = require('ini');
var lock = require('../util/lock');
var Prefs = require('../prefs');
var Promise = require('es6-promise').Promise;

var fsWrap = require('../util/fswrap');

// XXX: This should be 0600 not 0700
var RC_PERM_STRING = '0644';

var rc = module.exports = {};

rc.stat = function (rcPath) {

  return fsWrap.stat(rcPath).then(function (stat) {
    if (!stat.isFile()) {
      return Promise.reject(new Error('.arigatorc must be a file: ' + rcPath));
    }

    var fileMode = '0' + (stat.mode & parseInt('777', 8)).toString(8);

    if (fileMode !== RC_PERM_STRING) {
      var msg = 'rc file permission error: ' + rcPath + ' ' + fileMode +
        ' not ' + RC_PERM_STRING;
      return Promise.reject(new Error(msg));
    }

    return true;
  }).catch(function (err) {
    if (err && err.code === 'ENOENT') {
      return Promise.resolve(false);
    }

    return Promise.reject(err);
  });
};

rc.write = function (prefs) {
  if (!(prefs instanceof Prefs)) {
    return Promise.reject(new Error('prefs must be an instance of Prefs'));
  }

  return lock.wrap(prefs.path, function () {
    return rc.stat(prefs.path).then(function () {
      return fsWrap.write(prefs.path, ini.stringify(prefs.values));
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

      return fsWrap.read().then(function (contents) {
        return ini.parse(contents);
      });
    });
  });
};
