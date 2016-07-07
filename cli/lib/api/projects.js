'use strict';

var projects = exports;

projects.get = function (client, query) {
  return client.get({
    url: '/projects',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

projects.create = function (client, data) {
  return client.post({
    url: '/projects',
    json: {
      body: {
        org_id: data.org_id,
        name: data.name
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
