'use strict';

// remove once we have jsonschema
const emailValidator = require('email-validator');
const APP_NAME_REGEX = /^[a-zA-Z0-9\-_]{3,64}$/g;

// TODO: Re-use the json schema from the registry here :)
module.exports = {
  fullName: {
    type: 'input',
    message: 'Full Name',
    name: 'full_name',
    validate: (input) => {
      if (typeof input === 'string' && input.length >= 3 &&
          input.length <= 64) {
        return true;
      }

      return 'You must provide between 3 and 64 characters';
    }
  },
  email: {
    type: 'input',
    name: 'email',
    message: 'Email',
    validate: (input) => {
      if (emailValidator.validate(input)) {
        return true;
      }

      return 'You must provide a valid email address';
    }
  },
  password: {
    type: 'password',
    name: 'password',
    message: 'Password',
    validate: (input) => {
      if (typeof input === 'string' && input.length >= 12) {
        return true;
      }

      return 'You must provide a password greater than 12 characters in length';
    }
  },
  appName: {
    type: 'input',
    name: 'appName',
    message: 'Application Name',
    validate: (input) => {
      if (typeof input === 'string' && APP_NAME_REGEX.test(input)) {
        return true;
      }

      return 'Application names must match [a-zA-Z0-9\-_]{3,64}';
    }
  }
};
