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

users.profile = function (client, query, params) {
  return client.get({
    url: '/profiles/:username',
    qs: query || {},
    params: params || {}
  }).then(function (res) {
    return res.body;
  });
};

users.create = function (client, body) {
  return client.post({
    url: '/users',
    json: {
      id: utils.id('user'),
      version: 1,
      body: body
    }
  }).then(function (res) {
    return res.body;
  });
};

users.verify = function (client, body) {
  return client.post({
    url: '/users/verify',
    json: body
  });
};
