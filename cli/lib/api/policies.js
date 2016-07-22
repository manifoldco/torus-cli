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
      body: {
        org_id: data.org_id,
        version: 1 || data.version,
        previous: null || data.previous,
        type: USER_TYPE,
        policy: data.policy
      }
    }
  }).then(function (res) {
    return res.body;
  });
};
