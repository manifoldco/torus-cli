'use strict';

var list = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');

var SYSTEM_TYPE = 'system';

list.output = {};

var validator = validate.build({
  org: validate.slug
});

list.output.success = output.create(function (payload) {
  var org = payload.org;
  var teams = payload.teams;
  var memberships = _.groupBy(payload.memberships, 'body.team_id');

  var msg = ' teams in the ' + org.body.name + ' org (' + teams.length + ')\n';

  msg += ' ' + new Array(msg.length - 1).join('-') + '\n'; // underline
  msg += teams.map(function (team) {
    var systemTeam = (team.body.type === SYSTEM_TYPE);
    var m = '';

    if (memberships[team.id]) {
      m = '*';
    }

    m += team.body.name;
    m = _.padStart(m, m.length + 1);

    if (systemTeam) {
      m += ' [system]';
    }

    return m;
  }).join('\n');

  msg += '\n\n (*) member';

  console.log(msg);
});

list.output.failure = output.create(function () {
  console.log('Team retrieval failed, please try again.');
});

list.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    ctx.target.flags({
      org: ctx.option('org').value
    });

    var data = {
      org: ctx.target.org
    };

    if (!data.org) {
      throw new Error('--org is required.');
    }

    var errors = validator(data);
    if (errors.length > 0) {
      return reject(errors[0]);
    }

    var payload = {};
    return Promise.all([
      ctx.api.orgs.get({ name: data.org }),
      ctx.api.users.self()
    ])
    .then(function (results) {
      payload.org = results[0][0];
      payload.self = results[1];

      if (!payload.org) {
        throw new Error('org not found: ' + data.org);
      }

      if (!payload.self) {
        throw new Error('current user could not be retrieved');
      }

      return Promise.all([
        ctx.api.teams.get({ org_id: payload.org.id }),
        ctx.api.memberships.get({
          org_id: payload.org.id,
          owner_id: payload.self.id
        })
      ]);
    })
    .then(function (results) {
      payload.teams = results && results[0];
      payload.memberships = results && results[1];

      if (!_.isArray(payload.teams)) {
        return reject(new Error('could not find team(s)'));
      }

      if (!_.isArray(payload.memberships)) {
        return reject(new Error('could not find memberships'));
      }

      return resolve(payload);
    })
    .catch(reject);
  });
};
