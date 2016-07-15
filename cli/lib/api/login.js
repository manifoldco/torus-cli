'use strict';

var login = exports;
login.isNewAPI = true;

login.post = function (client, body) {
  return client.post({
    url: '/login',
    json: body
  }, login.isNewAPI).then(function (res) {
    return res.body;
  });
};
