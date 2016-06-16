'use strict';

var _ = require('lodash');
var path = require('path');

var PROPS = [
  'org',
  'project',
  'service',
  'environment'
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
