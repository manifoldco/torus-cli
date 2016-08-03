'use strict';

var invites = exports;

invites.list = function (client, query) {
  return client.get({
    url: '/org-invites',
    qs: query || {}
  }).then(function (res) {
    return res.body
  });
}
