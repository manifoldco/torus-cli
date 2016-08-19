'use strict';

var Promise = require('es6-promise').Promise;
var version = exports;

var NEW_API = true;
var OLD_API = false;

version.get = function (client) {
  return Promise.all([
    client.get({ url: '/version' }, OLD_API), // Registry version
    client.get({ url: '/version' }, NEW_API)  // Daemon version
  ]).then(function (responses) {
    return {
      registry: responses[0].body,
      daemon: responses[1].body
    };
  });
};
