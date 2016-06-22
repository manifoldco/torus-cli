'use strict';

var _ = require('lodash');
var path = require('path');

// Ordering of the properties matter
var PROPS = [
  'org',
  'project',
  'environment', // XXX How ot deal with env/service ??
  'service'
];

function defineProperty(obj, key) {
  Object.defineProperty(obj, key, {
    get: function () {
      return obj._values[key];
    },
    configurable: false,
    enumerable: true
  });
}

function Target(targetPath, context) {
  if (!_.isString(targetPath)) {
    throw new TypeError('Must provide a path string');
  }
  if (!path.isAbsolute(targetPath)) {
    throw new TypeError('Must provide an absolute path');
  }

  if (!_.isPlainObject(context)) {
    throw new TypeError('Must provide context object');
  }

  this.path = targetPath;
  this._values = {};

  var self = this;
  PROPS.forEach(function (prop) {
    var value = context[prop] || null;

    self._values[prop] = value;
    defineProperty(self, prop);
  });
}

module.exports = Target;

Target.prototype.flags = function (flags) {
  if (!_.isPlainObject(flags)) {
    throw new TypeError('Must provide context object');
  }

  // if the org or project is different then reset everything, it doesnt matter
  // if the env or service are different as they depend on project/org not each
  // other.
  if ((flags.org !== undefined || flags.project !== undefined) &&
      (flags.org !== this.org || flags.project !== this.project)) {
    if (flags.org !== undefined && flags.org !== this.org) {
      this._values.org = null;
    }

    this._values.project = null;
    this._values.service = null;
    this._values.environment = null;
  }

  var self = this;
  PROPS.forEach(function (prop) {
    if (flags[prop] === undefined) {
      return;
    }

    self._values[prop] = flags[prop];
  });
};

Target.prototype.context = function () {
  return _.clone(this._values);
};
