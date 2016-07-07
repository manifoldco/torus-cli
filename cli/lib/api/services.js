'use strict';

var services = exports;

services.get = function (client, query) {
  return client.get({
    url: '/services',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

services.create = function (client, data) {
  return client.post({
    url: '/services',
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
