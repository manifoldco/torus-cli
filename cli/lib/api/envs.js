'use strict';

var envs = exports;

envs.get = function (client, query) {
  return client.get({
    url: '/envs',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

envs.create = function (client, data) {
  return client.post({
    url: '/envs',
    json: {
      body: {
        org_id: data.org_id,
        project_id: data.project_id,
        name: data.name
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
