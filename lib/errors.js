'use strict';

class NotCancellableError extends Error {
  constructor (message, code) {
    arguments[0] = message || 'Command is not cancellable';
    arguments[1] = code || 'not_cancellable';
    Error.apply(this, arguments);
  }
}

class NotImplementedError extends Error {
  constructor (message, code) {
    arguments[0] = message || 'Command is not implemented';
    arguments[1] = code || 'not_implemened';
    Error.apply(this, arguments);
  }
}

exports.NotImplementedError = NotImplementedError;
exports.NotCancellableError = NotCancellableError;
