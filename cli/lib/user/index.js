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
  if (process.env.NODE_ENV !== 'arigato') {
    questions.splice(1, 0, INVITE_CODE);
  }

  passwordStage = questions.length - 1;
  return questions;
};

/**
 * Execute user creation from inputs
 */
user.execute = function (ctx, params) {
  var acceptOrgInvite = params && params.length === 3;

  var orgIndex = acceptOrgInvite ? 0 : -1;
  var codeIndex = acceptOrgInvite ? 2 : 1;
  var emailIndex = acceptOrgInvite ? 1 : 0;

  var defaults = {
    // Params is mixed type, fallback to array
    email: _.get(params, 'email', params[emailIndex]),
    code: _.get(params, 'code', params[codeIndex]),
    org: _.get(params, 'org', params[orgIndex])
  };

  // Create prompt from user questions
  var prompt = new Prompt({
    defaults: defaults,
    stages: user.questions
  });

  // Begin asking questions
  return prompt.start().then(function (userInput) {
    var opts = {
      params: params,
      defaults: defaults,
      userInput: userInput
    };
    // Create user object from input
    return user._create(ctx.api, opts).then(function (userObj) {
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
 * @param {object} opts
 */
user._create = function (api, opts) {
  var defaults = opts.defaults || {};
  var userInput = opts.userInput;
  var params = opts.params;
  var object = {
    username: userInput.username,
    name: userInput.name,
    email: userInput.email
  };

  var query = {};
  if (userInput.code || defaults.code) {
    query.code = userInput.code || defaults.code;
  }
  if (defaults.email || userInput.email) {
    query.email = defaults.email || userInput.email;
  }
  if (defaults.org) {
    query.org = defaults.org;
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
