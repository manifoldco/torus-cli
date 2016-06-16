'use strict';

var invite = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var client = require('../api/client').create();
var output = require('../cli/output');
var validate = require('../validate');

var validator = validate.build({
  team: validate.slug,
  username: validate.slug,
  org: validate.slug
});

invite.output = {};

invite.output.success = output.create(function (username) {
  console.log(username + ' has been invited to your org');
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
  /* eslint-disable consistent-return, no-shadow */
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

    client.auth(ctx.session.token);

    var getUser = {
      url: '/profiles/' + data.username
    };

    var getOrg = {
      url: '/orgs',
      qs: {
        name: data.org
      }
    };

    var user;
    var org;
    var team;

    Promise.all([
      client.get(getUser),
      client.get(getOrg)
    ])
    .then(function (results) {
      user = _.get(results, '[0].body[0]', null);
      if (!user) {
        throw new Error('user not found: ' + data.username);
      }

      org = _.get(results, '[1].body[0]', null);
      if (!org) {
        throw new Error('org not found: ' + data.org);
      }
    })
    .then(function () {
      var getTeam = {
        url: '/teams',
        qs: {
          org_id: org.id,
          name: data.team
        }
      };

      return client.get(getTeam).then(function (teams) {
        team = _.get(teams, 'body[0]', null);

        if (!team) {
          throw new Error('team not found: ' + data.team);
        }
      });
    })
    .then(function () {
      var postInvite = {
        url: '/memberships',
        json: {
          body: {
            org_id: org.id,
            owner_id: user.id,
            team_id: team.id
          }
        }
      };

      return client.post(postInvite);
    })
    .then(function () {
      resolve(data.username);
    })
    .catch(reject);
  });
};
