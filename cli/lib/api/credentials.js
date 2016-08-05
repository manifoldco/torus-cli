'use strict';

var credentials = exports;
credentials.isNewAPI = true;

var utils = require('common/utils');

credentials.get = function (client, query) {
  return client.get({
    url: '/credentials',
    qs: query || {}
  }, credentials.isNewAPI).then(function (res) {
    return res.body;
  });
};

credentials.create = function (client, data) {
  return client.post({
    url: '/credentials',
    json: {
      id: utils.id('credential'),
      version: 1,
      body: {
        name: data.name,
        project_id: data.project_id,
        org_id: data.org_id,
        pathexp: data.pathexp,
        version: data.version,
        previous: data.previous,
        value: data.value
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
