'use strict';

var policies = exports;

var utils = require('common/utils');

var USER_TYPE = 'user';

policies.get = function (client, query) {
  return client.get({
    url: '/policies',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

policies.create = function (client, data) {
  return client.post({
    url: '/policies',
    json: {
      id: utils.id('policy'),
      version: 1 || data.version,
      body: {
        org_id: data.org_id,
        previous: null || data.previous,
        type: USER_TYPE,
        policy: data.policy
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
