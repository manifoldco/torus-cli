'use strict';

class NotCancellableError extends Error {
  constructor (message, code) {
    message = message || 'Command is not cancellable';
    code = code || 'not_cancellable';
    super(message, code);
  }
}

class NotImplementedError extends Error {
  constructor (message, code) {
    message = message || 'Command is not implemented';
    code = code || 'not_implemened';
    super(message, code);
  }
}

exports.NotImplementedError = NotImplementedError;
exports.NotCancellableError = NotCancellableError;
