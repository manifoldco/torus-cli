'use strict';

var status = exports;

var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var client = require('../api/client').create();

status.output = {};

status.output.failure = output.create(function() {
  console.log('Could not identify your CLI, please log in and try again');
});

status.output.success = output.create(function(identity) {
  var user = identity.user;
  if (user && user.body) {
    console.log('Identity: ' + user.body.name + ' (' + user.body.email + ')');
  } else {
    console.log('Identity: Unauthenticated');
  }
});

status.execute = function(ctx) {
  var identity = {};
  return new Promise(function(resolve, reject) {
    if (!ctx.token) {
      identity.user = null;
      return resolve(identity);
    }

    client.auth(ctx.token);

    client.get({
      url: '/users/self',
    }).then(function(res) {
      var users = res.body || [];
      identity.user = users[0] || null;
      resolve(identity);
    }).catch(function(err) {
      switch (err.type) {
        case 'not_found':
          identity.user = null;
          resolve(identity);
        break;
        case 'unauthorized':
          identity.user = null;
          resolve(identity);
        break;
        default:
          reject(err);
        break;
      }
    });
  });
};
