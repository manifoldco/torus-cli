'use strict';

const pkg = require('../../package.json');

const version = exports;

const VERSION = pkg.version;

version.get = function () {
  return VERSION;
};
