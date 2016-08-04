'use strict';

var Promise = require('es6-promise').Promise;
var _ = require('lodash');
var Table = require('cli-table2');

var validate = require('../validate');
var output = require('../cli/output');

var invitesList = exports;

var validator = validate.build({
  org: validate.slug
});

invitesList.output = {};
invitesList.output.success = output.create(function (ctx, payload) {
  var orgName = payload.org.body.name;

  if (ctx.option('approved').value) {
    console.log(
      'Listing all approved invitations for the ' + orgName + ' org');
  } else {
    console.log(
      'Listing all pending and accepted invitations for the ' +
      orgName + ' org');
  }
  console.log();

  var table = new Table({
    head: [
      'Invite Email',
      'State',
      'Invited By',
      'Creation Date'
    ]
  });

  var userIndex = _.keyBy(payload.users, 'id');
  function formatEmail(invite) {
    var state = invite.body.state;
    if (['associated', 'accepted', 'approved'].indexOf(state) === -1) {
      return invite.body.email;
    }

    var user = userIndex[invite.body.invitee_id];
    if (!user) {
      throw new Error('Expected to find user: ' + invite.body.invitee_id);
    }

    return invite.body.email + ' (' + user.body.username + ')';
  }

  var rows = payload.invites.map(function (invite) {
    return [
      formatEmail(invite),
      invite.body.state,
      userIndex[invite.body.inviter_id].body.username,
      new Date(invite.body.created_at).toISOString()
    ];
  });

  if (rows.length === 0) {
    return console.log('No invites found.');
  }

  table.push.apply(table, rows);
  return console.log(table.toString());
});

invitesList.output.failure = output.create(function () {
  console.log('Retrieval of invites failed!');
});

invitesList.execute = function (ctx) {
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

    return ctx.api.orgs.get({ name: data.org }).then(function (orgs) {
      var org = orgs[0];
      if (!_.isObject(org)) {
        return reject(new Error('org not found: ' + data.org));
      }

      var state = (ctx.option('approved').value) ?
        ['approved'] : ['pending', 'associated', 'accepted'];

      var opts = {
        org_id: org.id,
        state: state
      };

      return ctx.api.invites.list(opts).then(function (invites) {
        var userIds = [];

        if (invites.length === 0) {
          return resolve({
            org: org,
            invites: [],
            users: {}
          });
        }

        invites.forEach(function (invite) {
          if (invite.body.invitee_id) {
            userIds.push(invite.body.invitee_id);
          }

          if (invite.body.approver_id) {
            userIds.push(invite.body.approver_id);
          }

          userIds.push(invite.body.inviter_id);
        });

        userIds = _.uniq(userIds);
        return ctx.api.profiles.list({ id: userIds }).then(function (users) {
          resolve({
            org: org,
            invites: invites,
            users: users
          });
        });
      });
    }).catch(reject);
  });
};
