'use strict';

var keypairs = exports;
keypairs.isNewAPI = true;

keypairs.list = function (client, query) {
  return client.get({
    url: '/keypairs',
    qs: query || {}
  }).then(function (res) {
    return res.body;
  });
};

keypairs.generate = function (client, body, progress) {
  return client.post({
    url: '/keypairs/generate',
    json: body
  }, keypairs.isNewAPI, progress).then(function (res) {
    return res.body;
  });
};
