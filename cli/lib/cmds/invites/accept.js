'use strict';

var Promise = require('es6-promise').Promise;

var Prompt = require('../../cli/prompt');
var Command = require('../../cli/command');
var flags = require('../../flags');
var invites = require('../../invites');
var userCreate = require('../../user/index');
var login = require('../../login/index');

function getUser(ctx) {
  return new Promise(function (resolve, reject) {
    ctx.api.users.self().then(function (user) {
      resolve(user);
    }).catch(function (err) {
      if (err.type === 'unauthorized') {
        return resolve();
      }

      return reject(err);
    });
  });
}

function signup(ctx, params) {
  return userCreate.execute(ctx, params).then(function (userInput) {
    return login.subcommand(ctx, userInput).then(function () {
      login.output.success();
    })
    .catch(function (err) {
      login.output.failure();

      throw err;
    });
  })
  .then(function () {
    return userCreate.finalize(ctx);
  })
  .then(function () {
    return ctx.api.invites.getByCode(params);
  })
  .catch(function (err) {
    userCreate.output.failure(err);

    throw err;
  });
}

function attemptLogin(ctx, params) {
  return login.execute(ctx).then(function () {
    login.output.success();

    return invites.accept.associate(ctx, params);
  }).catch(function (err) {
    login.output.failure();

    throw err;
  });
}

function becomeLoggedIn(ctx, params) {
  var prompt = new Prompt({
    stages: function () {
      return [
        [
          {
            type: 'list',
            name: 'step',
            message: 'Do you want to login or create an account?',
            choices: [
              'Signup',
              'Login'
            ]
          }
        ]
      ];
    }
  });

  return prompt.start().then(function (answers) {
    if (answers.step === 'Signup') {
      return signup(ctx, params);
    }

    return attemptLogin(ctx, params);
  });
}

var cmd = new Command(
  'invites:accept <email> <code>',
  'Accept an invitation to join an organization',
  'jeff@example.com 1a2b3c4d5e\n\n  Jeff will join the organization he was invited to',
  function (ctx) {
    return invites.accept.validate(ctx).then(function (params) {
      // Check if we're logged in; if we are then we can just go ahead and
      // assocaite directly with the invite.
      return getUser(ctx).then(function (user) {
        if (user) {
          return invites.accept.associate(ctx, params);
        }

        // If we're not logged in -- then we either need to log in or create an
        // account and login.
        return becomeLoggedIn(ctx, params);
      })
      .then(function (invite) {
        // Upload the keypair to the organization
        var opts = { org_id: invite.body.org_id };

        return ctx.api.keypairs.generate(opts);
      })
      .then(function () {
        // Accept that invite!
        return invites.accept.finalize(ctx, params);
      })
      .then(function () {
        invites.accept.output.success();

        return true;
      });
    })
    .catch(function (err) {
      invites.accept.output.failure();

      throw err;
    });
  }
);

flags.add(cmd, 'org');

module.exports = cmd;
