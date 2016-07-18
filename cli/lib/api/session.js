'use strict';

var session = exports;
session.isNewAPI = true;

session.get = function (client) {
  return client.get({
    url: '/session'
  }, session.isNewAPI).then(function (res) {
    return res.body;
  });
};
