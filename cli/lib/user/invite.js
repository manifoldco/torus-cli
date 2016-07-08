'use strict';

var invite = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');

var validator = validate.build({
  team: validate.slug,
  username: validate.slug,
  org: validate.slug
});

invite.output = {};

invite.output.success = output.create(function (results) {
  var username = results.profile.body.username;
  var orgName = results.org.body.name;
  var teamName = results.team.body.name;

  console.log(
    'You\'ve invited \'' + username + '\' to join your org ' +
    '(' + orgName + ') as a member of the ' + teamName + ' team.\n');
});

invite.output.failure = output.create(function () {
  console.error('Could not invite user, please try again');
});

/**
 * Perform invite based on ctx args
 *
 * @param {object} ctx - Command context
 */
invite.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      team: 'member',
      username: ctx.params[0],
      org: ctx.target.org
    };

    if (!data.org) {
      return reject(new Error('--org is required'));
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }


    var user;
    var org;
    var team;
    return Promise.all([
      ctx.api.users.profile({}, { username: data.username }),
      ctx.api.orgs.get({ name: data.org })
    ])
    .then(function (results) {
      user = _.get(results, '[0][0]', null);
      org = _.get(results, '[1][0]', null);

      if (!user) {
        throw new Error('user not found: ' + data.username);
      }

      if (!org) {
        throw new Error('org not found: ' + data.org);
      }
    })
    .then(function () {
      var qs = { org_id: org.id, name: data.team };
      return ctx.api.teams.get(qs).then(function (teams) {
        team = teams[0];
        if (!team) {
          throw new Error('team not found: ' + data.team);
        }
      });
    })
    .then(function () {
      return ctx.api.memberships.create({
        org_id: org.id,
        owner_id: user.id,
        team_id: team.id
      });
    })
    .then(function () {
      resolve({
        profile: user,
        org: org,
        team: team
      });
    })
    .catch(reject);
  });
};
