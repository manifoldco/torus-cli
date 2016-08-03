'use strict';

var _ = require('lodash');
var utils = require('common/utils');

var invites = exports;

invites.create = function (client, data) {
  var orgInvite = {
    id: utils.id('org_invite'),
    version: 1,
    body: {
      org_id: data.org.id,
      inviter_id: data.user.id,
      pending_teams: _.map(data.pending_teams, 'id'),
      email: data.email,
      invitee_id: null,
      approver_id: null,
      created_at: new Date().toISOString(),
      accepted_at: null,
      approved_at: null
    }
  };


  return client.post({
    url: '/org-invites',
    json: orgInvite
  }).then(function (res) {
    return res.body;
  });
};

invites.list = function (client, query) {
  return client.get({
    url: '/org-invites',
    qs: query || {}
  }).then(function (res) {
    return res.body
  });
}
