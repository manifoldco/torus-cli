'use strict';

var status = exports;

var Promise = require('es6-promise').Promise;

var Target = require('../context/target');
var output = require('../cli/output');

status.output = {};

status.output.failure = output.create(function () {
  console.log('Error determining status');
});

status.output.success = output.create(function (ctx, state) {
  var programName = ctx.program.name;

  console.log('Current Working Context:\n');
  console.log(
    'Identity: ' + state.user.body.name + ' (' + state.user.body.email + ')');
  console.log('Username: ' + state.user.body.username);

  if (state.target.disabled()) {
    return console.log('\nContext is disabled for your CLI,' +
      ' use \'' + programName + ' prefs\' to enabled it');
  }

  if (state.target.org === null) {
    return console.log('\nYou are not inside a linked working directory,' +
      ' use \'' + programName + ' link\' to link your project');
  }

  var service = state.target.service || '*';

  console.log('Org: ' + state.target.org);
  console.log('Project: ' + state.target.project);
  console.log('Environment: ' + state.target.environment);
  console.log('Service: ' + service);
  console.log('Instance: *');

  return console.log('\nCredential Path: /' + [
    state.target.org,
    state.target.project,
    state.target.environment,
    service,
    state.user.body.username,
    '*'
  ].join('/'));
});

status.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.target instanceof Target)) {
      throw new Error('Target must be on the context object');
    }

    return ctx.api.users.self().then(function (user) {
      if (!user) {
        throw new Error('Invalid response returned from the API');
      }

      resolve({
        user: user,
        target: ctx.target
      });
    }).catch(reject);
  });
};
