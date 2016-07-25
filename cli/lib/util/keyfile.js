'use strict';

var _ = require('lodash');
var base64url = require('base64url');
var path = require('path');

var fsWrap = require('./fswrap');

var keyFile = exports;

var KEYFILE_PERM_STRING = '0644';

/**
 * Returns true or false depending on whether or not the given key file path is
 * a valid arigato public key distribution file.
 */
keyFile.validate = function (keyFilePath) {
  keyFilePath = path.resolve(process.cwd(), keyFilePath);

  return fsWrap.stat(keyFilePath).then(function (stat) {
    if (!stat.isFile()) {
      return 'must be a file';
    }

    var fileMode = '0' + (stat.mode & parseInt('777', 8)).toString(8);
    if (fileMode !== KEYFILE_PERM_STRING) {
      return 'invalid file permissions, must be: ' + KEYFILE_PERM_STRING;
    }

    return fsWrap.read(keyFilePath);
  }).then(function (data) {
    var contents = JSON.parse(data);
    if (!_.isPlainObject(contents) ||
        !_.isEqual(Object.keys(contents).sort(), ['public_key']) ||
        base64url.toBuffer(contents.public_key).length !== 32) {
      return 'invalid file contents; missing key or invalid key encoding';
    }

    return true;
  });
};
