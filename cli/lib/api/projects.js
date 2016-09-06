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
