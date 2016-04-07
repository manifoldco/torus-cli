'use strict';

function Option (flags, description, defaultValue) {

  this.flags = flags;
  this.required = (flags.indexOf('<') > -1);
  this.optional = (flags.indexOf('[') > -1);
  this.hasParam = (this.required || this.optional);
  this.bool = (flags.indexOf('--no-') > -1 ||
               !(this.required || this.optional));
  this.defaultValue = defaultValue;
  this._value = undefined;


  if (this.required && this.optional &&
      (flags.indexOf('[') < flags.indexOf('<'))) {
    throw new Error('Required must come before optional params: '+flags);
  }

  flags = flags.split(/[ ,|]+/);
  if (flags.length > 1 && !/^[[<]/.test(flags[1])) {
    this.short = flags.shift();
  }
  this.long = flags.shift();
  this.description = description || '';

  // If option is --no-* then set to true
  if (this.bool) {
    this.defaultValue = (this.long.indexOf('-no-') > -1) ? true : false;
  }
}

Option.prototype.name = function () {
  return this.long.replace('--','').replace('no-', '');
};

Option.prototype.shortcut = function () {
  return this.short.replace(/^\-/, '');
}

Option.prototype.evaluate = function (ctx, args) {
  ctx.props[this.name()] = this;
  this.ctx = ctx;

  var value = args[this.shortcut()] || args[this.name()];
  if (value === undefined && this.defaultValue !== undefined) {
    value = this.defaultValue;
  }

  value = (value === undefined && this.defaultValue !== undefined ) ?
    this.defaultValue : value;

  this.value = value;
};

module.exports = Option;
