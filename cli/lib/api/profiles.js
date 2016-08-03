'use strict';

var profiles = exports;

profiles.list = function (client, query) {
  return client.get({
    url: '/profiles',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

profiles.get = function (client, query, params) {
  return client.get({
    url: '/profiles/:username',
    qs: query || {},
    params: params || {}
  }).then(function (res) {
    return res.body;
  });
};
 
