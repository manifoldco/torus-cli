'use strict';

var FLAG_REGEX =
  /^\-{1}[a-z]{1}, \-\-[a-z\-]{2,16}\s*(\[[a-z]+\]|<[a-z]+>)?$/;

/**
 * Represents an Option supported by a Command.
 *
 * @param {String} Flags string of flags supporting the flag grammar
 * @param {String} description defaults ot an empty string
 * @param {*} defaultValue default value
 *
 * Flag Grammar:
 *
 *  -p, --pretty [name]   parameter has optional value provided
 *  -p, --pretty <name>   parameter is required
 *  -n, --no-x            flag that defauls to true if not provided
 *  -i, --invite          flag that defaults to false if not provided
 */
function Option(flags, description, defaultValue) {
  if (typeof flags !== 'string') {
    throw new TypeError('flags must be a string');
  }

  if (!flags.match(FLAG_REGEX)) {
    throw new Error('Flags do not match the regex');
  }

  this.flags = flags;
  this.required = (flags.indexOf('<') > -1);
  this.optional = (flags.indexOf('[') > -1);
  this.hasParam = (this.required || this.optional);
  this.bool = (flags.indexOf('--no-') > -1 ||
               !(this.required || this.optional));
  this.defaultValue = defaultValue;
  this._value = undefined;


  flags = flags.split(/[ ,|]+/);
  if (flags.length > 1 && !/^[[<]/.test(flags[1])) {
    this.short = flags.shift();
  }
  this.long = flags.shift();
  this.description = description || '';

  // If option is --no-* then set to true
  if (this.bool) {
    this.defaultValue = this.long.indexOf('-no-') > -1;
  }
}

Option.prototype.name = function () {
  return this.long.replace('--', '').replace('no-', '');
};

Option.prototype.shortcut = function () {
  return this.short.replace(/^\-/, '');
};

Option.prototype.evaluate = function (ctx, args) {
  ctx.options[this.name()] = this;
  this.ctx = ctx;

  var value = args[this.shortcut()] || args[this.name()];
  if (value === undefined && this.defaultValue !== undefined) {
    value = this.defaultValue;
  }

  value = (value === undefined && this.defaultValue !== undefined) ?
    this.defaultValue : value;

  this.value = value;
};

module.exports = Option;
