'use strict';

var util = require('util');

var wrap = require('./wrap');
var Option = require('./option');
var Runnable = require('./runnable');

var COMMAND_REGEX =
  /^[a-z]{2,16}(:[a-z]{2,16})?\s*(<[a-z]{2,16}> ?)*\s*(\[[a-z]{2,16}\] ?)*$/;

function Command(usage, description, handler) {
  Runnable.call(this);

  if (typeof usage !== 'string') {
    throw new TypeError('Usage must be a string');
  }
  if (!usage.match(COMMAND_REGEX)) {
    throw new Error('Usage does not match regex');
  }

  if (typeof handler !== 'function' &&
      (typeof handler !== 'object' || typeof handler.run !== 'function')) {
    throw new TypeError('Handler must be a function or object with run method');
  }

  this._handler = handler;
  this.usage = usage;
  this.description = description || '';
  this.slug = usage.split(' ')[0];

  var parts = this.slug.split(':');
  this.group = (parts.length > 0) ? parts[0] : this.slug;
  this.subpath = (parts.length > 1) ? parts[1] : this.group;

  this.options = [];

  return this;
}
util.inherits(Command, Runnable);
module.exports = Command;

Command.prototype.option = function (usage, description, defaultValue) {
  if (usage instanceof Option) {
    return this.options.push(usage);
  }

  this.options.push(new Option(usage, description, defaultValue));
  return this;
};

Command.prototype.run = function (ctx) {
  var self = this;

  function call() {
    if (typeof self._handler === 'function') {
      return wrap(self._handler.bind(self._handler, ctx));
    }

    return wrap(self._handler.run.bind(self._handler, ctx));
  }

  return self.runHooks('pre', ctx).then(function (success) {
    if (success === false) {
      return false;
    }

    return call().then(function () {
      return self.runHooks('post', ctx);
    });
  });
};
