'use strict';

var teams = exports;

teams.get = function (client, query) {
  return client.get({
    url: '/teams',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

teams.create = function (client, data) {
  return client.post({
    url: '/teams',
    json: {
      body: {
        org_id: data.org_id,
        name: data.name,
        type: data.type
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
