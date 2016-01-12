'use strict';

// remove once we have jsonschema
const emailValidator = require('email-validator');

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
  }
};
