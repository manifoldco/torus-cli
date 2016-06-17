'use strict';

var invite = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var client = require('../api/client').create();
var output = require('../cli/output');
var validate = require('../validate');

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
  return new Promise(function (resolve, reject) {
    var data = {
      team: 'member',
      username: ctx.params[0],
      org: ctx.option('org').value
    };

    if (!data.org) {
      throw new Error('--org is required');
    }

    if (validate.slug(data.org) instanceof Error) {
      throw new Error('invalid org name supplied');
    }

    if (!data.username) {
      throw new Error('username is required');
    }

    if (validate.slug(data.username) instanceof Error) {
      throw new Error('invalid username supplied');
    }

    client.auth(ctx.session.token);

    var getUser = {
      url: '/profiles/' + data.username
    };

    var getTeam = {
      url: '/teams',
      qs: {
        name: data.team
      }
    };

    var getOrg = {
      url: '/orgs',
      qs: {
        name: data.org
      }
    };

    Promise.all([
      client.get(getUser),
      client.get(getTeam),
      client.get(getOrg)
    ])
    .then(function (results) {
      var user = _.get(results, '[0].body[0]', null);
      if (!user) {
        throw new Error('user not found: ' + data.username);
      }

      var team = _.get(results, '[1].body[0]', null);
      if (!team) {
        throw new Error('team not found: ' + data.team);
      }

      var org = _.get(results, '[2].body[0]', null);
      if (!org) {
        throw new Error('org not found: ' + data.org);
      }

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
