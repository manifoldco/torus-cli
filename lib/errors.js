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

class RegistryError extends Error {
  constructor (message, code) {
    message = message || 'Encountered an error while speaking to registry';
    code = code || 'registry_error';
    super(message, code);
  }
}

class AlreadyAuthenticatedError extends Error {
  constructor (message, code) {
    message = message || 'You\'ve already authenticated';
    code = code || 'already_authenticated';

    super(message, code);
  }
}

exports.NotImplementedError = NotImplementedError;
exports.NotCancellableError = NotCancellableError;
exports.RegistryError = RegistryError;
exports.AlreadyAuthenticatedError = AlreadyAuthenticatedError;
