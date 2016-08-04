'use strict';

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var validate = require('../validate');
var output = require('../cli/output');

var validator = validate.build({
  email: validate.email,
  org: validate.slug,
  team: validate.slug
});

var inviteSend = exports;

inviteSend.output = {};
inviteSend.output.success = output.create(function (payload) {
  var orgName = payload.org.body.name;
  var email = payload.invite.body.email;

  console.log('Invitation to join the ' + orgName +
              ' organization has been sent to ' + email + '.');
  console.log();
  console.log('They will be added to the following teams once their ' +
              'invite has been confirmed:\n');

  payload.teams.forEach(function (team) {
    console.log('\t' + team.body.name);
  });

  console.log();
  console.log('They will receive an e-mail with instructions.');
});

inviteSend.output.failure = output.create(function () {
  console.log('Could not send invitation to org, please try again.');
});

inviteSend.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    // Mutate teamValue into array to deal with singular and multiple uses of
    // --team flag
    var teamValue = ctx.option('team').value || [];
    teamValue = Array.isArray(teamValue) ? teamValue : [teamValue];

    // Member is a system team that all users must belong too inorder to be
    // true members of an organization.
    teamValue.push('member');
    teamValue = _.uniq(teamValue); // remove duplicates

    var data = {
      org: ctx.target.org,
      email: ctx.params[0],
      team: teamValue
    };

    if (!data.email) {
      throw new Error('email parameter is required');
    }

    if (!data.org) {
      throw new Error('--org is required.');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    var user;
    var org;
    var pendingTeams;
    return Promise.all([
      ctx.api.users.self(),
      ctx.api.orgs.get({ name: data.org })
    ]).then(function (results) {
      user = results[0];
      org = results[1][0];

      if (!user) {
        throw new Error('user not found');
      }

      if (!org) {
        throw new Error('org not found: ' + data.org);
      }
    })
    .then(function () {
      return ctx.api.teams.get({ org_id: org.id });
    })
    .then(function (teams) {
      // Validate that the given team names are right.
      var teamNameIndex = _.keyBy(teams, 'body.name');
      var unknownTeams = data.team.filter(function (name) {
        return teamNameIndex[name] === null ||
          teamNameIndex[name] === undefined;
      });

      if (unknownTeams.length > 0) {
        throw new Error('Unknown team name: ' + unknownTeams[0]);
      }

      pendingTeams = data.team.map(function (name) {
        return teamNameIndex[name];
      });

      return ctx.api.invites.create({
        user: user,
        org: org,
        pending_teams: pendingTeams,
        email: data.email
      });
    })
    .then(function (invite) {
      return resolve({
        org: org,
        teams: pendingTeams,
        invite: invite
      });
    })
    .catch(reject);
  });
};
