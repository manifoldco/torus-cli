'use strict';

var invites = exports;

invites.list = function (client, query) {
  return client.get({
    url: '/org-invites',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

invites.getByCode = function (client, query) {
  return client.get({
    url: '/org-invites/code',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

invites.associate = function (client, body) {
  return client.post({
    url: '/org-invites/associate',
    json: body || {}
  }).then(function (res) {
    return res.body;
  });
};

invites.accept = function (client, body) {
  return client.post({
    url: '/org-invites/accept',
    json: body || {}
  }).then(function (res) {
    return res.body;
  });
};

invites.approve = function (client, query, params, progress) {
  return client.post({
    url: '/org-invites/:id/approve',
    qs: query || {},
    params: params || {}
  }, true, progress).then(function (res) {
    return res.body;
  });
};
