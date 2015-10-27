'use strict';

var _ = require('lodash');

function prependArguments (args, param) {
  args = Array.prototype.slice.apply(args);
  args.unshift(param);

  return args;
}

function Logger (name) {
  this.name = name;
}
module.exports.Logger = Logger;

Logger.prototype.error = function () {
  var args = prependArguments(arguments, 'error');
  this.print.apply(this, args);
};

Logger.prototype.warn = function () {
  var args = prependArguments(arguments, 'warn');
  this.print.apply(this, args);
};

Logger.prototype.debug = function () {
  arguments.unshift('debug');
  this.print.apply(this, arguments);
};

Logger.prototype.print = function () {
  console.log.apply(this, arguments);
};

var loggers = {};
module.exports.get = function (logger) {

  if (!logger || !_.isString(logger)) {
    throw new Error(`Cannot create logger: invalid string: ${logger}`);
  }
  logger = logger.toLowerCase().split('.')[0];

  if (!loggers[logger]) {
    loggers[logger] = new Logger(logger);
  }

  return loggers[logger];
};
