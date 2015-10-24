'use strict';

var _ = require('lodash');

function Logger (name) {
  this.name = name;
}
module.exports.Logger = Logger;

Logger.prototype.error = function () {
  arguments.unshift('error');
  this.log.apply(this, arguments);
};

Logger.prototype.warn = function () {
  arguments.unshift('warn');
  this.log.apply(this, arguments);
};

Logger.prototype.debug = function () {
  arguments.unshift('debug');
  this.log.apply(this, arguments);
};

Logger.prototype.log = function () {
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
