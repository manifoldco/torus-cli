'use strict';

var status = exports;

var Promise = require('es6-promise').Promise;

var Target = require('../context/target');
var output = require('../cli/output');
var client = require('../api/client').create();

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

  if (state.target.org === null) {
    return console.log('\nYou are not inside a linked working directory,' +
      ' use \'' + programName + ' init\' to link your project');
  }

  console.log('Org: ' + state.target.org);
  console.log('Project: ' + state.target.project);
  console.log('Environment: dev-' + state.user.body.username);
  console.log('Service: ' + state.target.service);
  console.log('Instance: 1 (Default)');

  return console.log('\nCredential Path: /' + [
    state.target.org,
    state.target.project,
    'dev-' + state.user.body.username,
    state.target.service,
    state.user.body.username,
    '1'
  ].join('/'));
});

status.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    if (!(ctx.target instanceof Target)) {
      throw new Error('Target must be on the context object');
    }

    client.auth(ctx.session.token);

    return client.get({
      url: '/users/self'
    }).then(function (res) {
      var user = res && res.body && res.body[0];
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
