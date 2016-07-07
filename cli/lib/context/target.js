'use strict';

var _ = require('lodash');
var path = require('path');

// Ordering of the properties matter
var PROPS = [
  'org',
  'project',
  'environment', // XXX How to deal with env/service ??
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

function Target(map) {
  if (!_.isPlainObject(map)) {
    throw new TypeError('Must provide map object');
  }
  if (!_.isString(map.path)) {
    throw new TypeError('Must provide map.path string');
  }
  if (!path.isAbsolute(map.path)) {
    throw new Error('Must provide an absolute map.path');
  }
  if (!_.isPlainObject(map.context) && !_.isNull(map.context)) {
    throw new TypeError('Must provide map.context object or null');
  }

  this._map = map;
  this._values = {};

  var self = this;
  var context = map.context || {};
  PROPS.forEach(function (prop) {
    var value = context[prop] || null;

    self._values[prop] = value;
    defineProperty(self, prop);
  });
}

module.exports = Target;

Target.prototype.exists = function () {
  return _.isPlainObject(this._map.context);
};

Target.prototype.path = function () {
  return this._map.path;
};

Target.prototype.flags = function (flags) {
  if (!_.isPlainObject(flags)) {
    throw new TypeError('Must provide context object');
  }

  // if the org or project is different then reset everything, it doesnt matter
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
