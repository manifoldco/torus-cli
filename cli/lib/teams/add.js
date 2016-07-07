'use strict';

var add = exports;

var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');

add.output = {};

var validator = validate.build({
  username: validate.slug,
  team: validate.slug,
  org: validate.slug
});

add.output.success = output.create(function (payload) {
  var team = payload.team.body.name;
  var username = payload.user.body.username;

  console.log(username + ' has been added to the ' + team + ' team.');
});

add.output.failure = output.create(function () {
  console.log('Failed to add the user to the team, please try again');
});

add.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      org: ctx.target.org,
      username: ctx.params[0],
      team: ctx.params[1]
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }

    if (!data.username || !data.team) {
      throw new Error('username and team are required.');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    var membershipData = {};
    var payload = {};
    return Promise.all([
      ctx.api.orgs.get({ name: data.org }),
      ctx.api.users.profile({}, { username: data.username })
    ])
    .then(function (result) {
      payload.org = result[0] && result[0][0];
      payload.user = result[1] && result[1][0];

      if (!payload.org) {
        throw new Error('org not found: ' + data.org);
      }

      if (!payload.user) {
        throw new Error('user not found: ' + data.username);
      }

      membershipData.org_id = payload.org.id;
      membershipData.owner_id = payload.user.id;

      return ctx.api.teams.get({
        org_id: payload.org.id,
        name: data.team
      });
    })
    .then(function (results) {
      payload.team = results && results[0];

      if (!payload.team) {
        throw new Error('team not found: ' + data.team);
      }

      membershipData.team_id = payload.team.id;
      return ctx.api.memberships.create(membershipData);
    })
    .then(function () {
      resolve(payload);
    })
    .catch(reject);
  });
};
