'use strict';

var user = exports;

var _ = require('lodash');

var Prompt = require('../cli/prompt');
var validate = require('../validate');
var output = require('../cli/output');
var userCrypto = require('common/crypto/user');

user.output = {};

user.output.success = output.create(function () {
  console.log('Your account has been created!');
});

user.output.failure = function (err) {
  var message = err.message;
  var messages = Array.isArray(message) ? message : [message];

  output.create.call(null, function () {
    switch (err.type) {
      // TODO: Graceful re-start of prompt for invalid input
      case 'invalid_request':
        if (messages.indexOf('resource exists') > -1) {
          console.error('Email address in use, please try again');
        } else {
          console.error('Whoops! Something went wrong. Please try again');
        }
        break;
      default:
        console.error('Signup failed, please try again');
        break;
    }
  });
};

/**
 * Signup prompt questions
 */
user.questions = function () {
  var self = this;
  var passwordStage;

  var INVITE_CODE = [
    {
      name: 'code',
      message: 'Invite code',
      validate: validate.inviteCode
    }
  ];

  var questions = [
    [
      {
        name: 'name',
        message: 'Full name',
        validate: validate.name
      },
      {
        name: 'username',
        message: 'Username',
        validate: validate.slug
      }
    ],
    [
      {
        name: 'email',
        message: 'Email',
        validate: validate.email
      }
    ],
    [
      {
        type: 'password',
        name: 'passphrase',
        message: 'Passphrase',
        validate: validate.passphrase,
        retryMessage: 'Passwords did not match. Please re-enter your password'
      },
      {
        type: 'password',
        name: 'confirm_passphrase',
        message: 'Confirm Passphrase',
        validate: function (input, answers) {
          var error = 'Passphrase must match';
          var failed = input !== answers.passphrase;
          // When failed, tell prompt which stage failed to recurse
          // if the maximum attempts has been reached
          return failed ? self.failed(passwordStage, error) : true;
        }
      }
    ]
  ];

  // All non-dev environments require invite code
  if (process.env.NODE_ENV !== 'development') {
    questions.splice(1, 0, INVITE_CODE);
  }

  passwordStage = questions.length - 1;
  return questions;
};

/**
 * Execute user creation from inputs
 */
user.execute = function (ctx, params) {
  var defaults;
  if (_.isArray(params)) {
    defaults = {
      email: params[0],
      code: params[1]
    };
  } else if (_.isPlainObject(params)) {
    defaults = {
      email: params.email,
      code: params.code
    };
  }
  // Create prompt from user questions
  var prompt = new Prompt({
    defaults: defaults,
    stages: user.questions
  });

  // Begin asking questions
  return prompt.start().then(function (userInput) {
    // Create user object from input
    return user._create(ctx.api, userInput, params).then(function (userObj) {
      return {
        inputs: userInput,
        user: userObj
      };
    });
  });
};

/**
 * Finalize user signup *after* the user has logged in
 */
user.finalize = function (ctx) {
  return ctx.api.orgs.get().then(function (orgs) {
    return ctx.api.keypairs.generate({ org_id: orgs[0].id });
  });
};

/**
 * Create user object from inputs
 *
 * @param {object} api api client
 * @param {object} userInput
 */
user._create = function (api, userInput) {
  var object = {
    username: userInput.username,
    name: userInput.name,
    email: userInput.email
  };

  var query = {};
  if (userInput.code) {
    query.code = userInput.code;
  }

  // Encrypt the password, generate the master key
  return userCrypto.encryptPasswordObject(userInput.passphrase)
    .then(function (result) {
      // Append the master and password objects to body
      object = _.extend(object, result);
    }).then(function () {
      return api.users.create(object, query);
    });
};
