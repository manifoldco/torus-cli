'use strict';

var memberships = exports;

memberships.get = function (client, query) {
  return client.get({
    url: '/memberships',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

memberships.delete = function (client, query, params) {
  return client.delete({
    url: '/memberships/:id',
    qs: query || {},
    params: params || {}
  }).then(function (res) {
    return res.body;
  });
};

memberships.create = function (client, data) {
  return client.post({
    url: '/memberships',
    json: {
      body: {
        org_id: data.org_id,
        owner_id: data.owner_id,
        team_id: data.team_id
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
