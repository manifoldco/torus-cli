'use strict';

var keypairs = exports;
keypairs.isNewAPI = true;

keypairs.generate = function (client, body) {
  return client.post({
    url: '/keypairs/generate',
    json: body
  }, keypairs.isNewAPI).then(function (res) {
    return res.body;
  });
};
