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

class NotAuthenticatedError extends Error {
  constructor (message, code) {
    message = message || 'You are not authenticated!';
    code = code || 'not_authenticated';

    super(message, code);
  }
}

class ArigatoConfigError extends Error {
  constructor (message, code) {
    message = message || 'Configuration Error in arigato.yaml';
    code = code || 'invalid_arigato_yaml_error';

    super(message, code);
  }
}

class ValidationError extends Error {
  constructor (message, code) {
    message = message || 'User input did not pass validation';
    code = code || 'validation_error';

    super(message, code);
  }
}

exports.NotImplementedError = NotImplementedError;
exports.NotCancellableError = NotCancellableError;
exports.RegistryError = RegistryError;
exports.AlreadyAuthenticatedError = AlreadyAuthenticatedError;
exports.NotAuthenticatedError = NotAuthenticatedError;
exports.ArigatoConfigError = ArigatoConfigError;
exports.ValidationError = ValidationError;
