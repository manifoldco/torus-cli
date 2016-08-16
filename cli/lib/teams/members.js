'use strict';

var members = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');

var validator = validate.build({
  name: validate.slug,
  org: validate.slug
});

members.output = {};

members.output.success = output.create(function (payload) {
  var team = payload.team;
  var total = payload.users.length;
  var msg = ' members of the ' + team.body.name + ' team (' + total + ')\n';
  msg += ' ' + new Array(msg.length - 1).join('-') + '\n';
  var includesMe = false;
  msg += payload.users.map(function (user) {
    var star = '';
    if (user.body.username === payload.self.body.username) {
      star = '*';
      includesMe = true;
    }
    return ' ' + star + user.body.name + ' [' + user.body.username + ']';
  }).join('\n');
  if (includesMe) {
    msg += '\n\n(*) you';
  }
  console.log(msg);
});

members.output.failure = output.create(function () {
  console.log('Team retrieval failed, please try again.');
});

members.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      org: ctx.target.org,
      name: ctx.params[0]
    };

    if (!data.name) {
      throw new Error('team name required.');
    }

    if (!data.org) {
      throw new Error('--org is required.');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    var org;
    var payload = {};
    return ctx.api.users.self()
    .then(function (self) {
      payload.self = self;
      return ctx.api.orgs.get({
        name: data.org
      });
    })
    .then(function (orgs) {
      org = orgs && orgs[0];
      if (!org) {
        throw new Error('unknown org: ' + data.org);
      }
      payload.org = org;
      return ctx.api.teams.get({
        org_id: org.id,
        name: data.name
      });
    })
    .then(function (teams) {
      var team = teams && teams[0];
      if (!team) {
        throw new Error('unknown team: ' + data.name);
      }

      payload.team = team;
      return ctx.api.memberships.get({
        team_id: team.id,
        org_id: org.id
      });
    })
    .then(function (memberships) {
      if (memberships.length < 1) {
        return [];
      }
      return ctx.api.profiles.list({
        id: _.map(memberships, function (membership) {
          return membership.body.owner_id;
        })
      });
    })
    .then(function (users) {
      payload.users = users;
      return resolve(payload);
    })
    .catch(reject);
  });
};
