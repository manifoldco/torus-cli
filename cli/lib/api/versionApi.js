'use strict';

var version = exports;
version.isNewAPI = true;

version.get = function (client, query) {
  return client.get({
    url: '/version',
    qs: query || {}
  }, version.isNewAPI).then(function (res) {
    return res.body;
  });
};
