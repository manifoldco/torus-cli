'use strict';

var logout = exports;
logout.isNewAPI = true;

logout.post = function (client, query) {
  return client.post({
    url: '/logout',
    qs: query || {}
  }, logout.isNewAPI).then(function (res) {
    return res.body;
  });
};
