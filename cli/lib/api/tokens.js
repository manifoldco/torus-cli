'use strict';

var tokens = exports;

tokens.create = function (client, body) {
  return client.post({
    url: '/tokens',
    json: body
  }).then(function (res) {
    return res.body;
  });
};

tokens.remove = function (client, query, params) {
  return client.delete({
    url: '/tokens/:auth_token',
    qs: query || {},
    params: params || {}
  }).then(function () {
    return {};
  });
};
