'use strict';

var users = exports;

var utils = require('common/utils');

users.self = function (client) {
  return client.get({
    url: '/users/self'
  }).then(function (res) {
    return res.body;
  });
};

users.create = function (client, body, query) {
  return client.post({
    url: '/users',
    json: {
      id: utils.id('user'),
      version: 1,
      body: body
    },
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};
