'use strict';

var list = exports;

var _ = require('lodash');
var Promise = require('es6-promise').Promise;

var output = require('../cli/output');
var validate = require('../validate');
var client = require('../api/client').create();

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
    client.auth(ctx.session.token);

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
      client.get({
        url: '/orgs',
        qs: { name: data.org }
      }),
      client.get({
        url: '/users/self'
      })
    ])
    .then(function (results) {
      payload.org = results[0].body && results[0].body[0];
      payload.self = results[1].body && results[1].body[0];

      if (!payload.org) {
        throw new Error('org not found: ' + data.org);
      }

      if (!payload.self) {
        throw new Error('current user could not be retrieved');
      }

      return Promise.all([
        client.get({
          url: '/teams',
          qs: {
            org_id: payload.org.id
          }
        }),
        client.get({
          url: '/memberships',
          qs: {
            org_id: payload.org.id,
            owner_id: payload.self.id
          }
        })
      ]);
    })
    .then(function (results) {
      payload.teams = results[0].body;
      payload.memberships = results[1].body;

      if (!payload.teams || !_.isArray(payload.teams)) {
        return reject(new Error('could not find team(s)'));
      }

      if (!payload.memberships || !_.isArray(payload.memberships)) {
        return reject(new Error('could not find memberships'));
      }

      return resolve(payload);
    })
    .catch(reject);
  });
};
