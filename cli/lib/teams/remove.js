'use strict';

var remove = exports;

var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');

remove.output = {};

var validator = validate.build({
  username: validate.slug,
  team: validate.slug,
  org: validate.slug
});

remove.output.success = output.create(function (payload) {
  var team = payload.team.body.name;
  var username = payload.user.body.username;

  console.log(username + ' has been removed from the ' + team + ' team.');
});

remove.output.failure = output.create(function () {
  console.log('Failed to remove the user from the team, please try again');
});

remove.execute = function (ctx) {
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

    var payload = {};
    return Promise.all([
      ctx.api.orgs.get({ name: data.org }),
      ctx.api.profiles.get({}, { username: data.username })
    ])
    .then(function (result) {
      payload.org = result[0] && result[0][0];
      if (!payload.org) {
        throw new Error('org not found: ' + data.org);
      }

      payload.user = result[1];
      if (!payload.user) {
        throw new Error('user not found: ' + data.username);
      }

      return ctx.api.teams.get({
        org_id: payload.org.id,
        name: data.team
      }).then(function (results) {
        return results && results[0];
      });
    })
    .then(function (team) {
      if (!team) {
        throw new Error('team not found: ' + data.team);
      }

      payload.team = team;
      return ctx.api.memberships.get({
        org_id: payload.org.id,
        team_id: payload.team.id,
        owner_id: payload.user.id
      });
    })
    .then(function (memberships) {
      payload.membership = memberships && memberships[0];
      if (!payload.membership) {
        throw new Error('user is not a member of team: ' + data.team);
      }

      return ctx.api.memberships.delete({
        id: payload.membership.id
      });
    })
    .then(function () {
      resolve(payload);
    })
    .catch(reject);
  });
};
