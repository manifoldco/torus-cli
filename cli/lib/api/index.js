/* eslint global-require: "off" */
'use strict';

var _ = require('lodash');

var api = exports;

api.modules = {
  users: require('./users'),
  tokens: require('./tokens'),
  orgs: require('./orgs'),
  teams: require('./teams'),
  memberships: require('./memberships'),
  projects: require('./projects'),
  envs: require('./envs'),
  services: require('./services'),
  credentials: require('./credentials')
};

api.Client = require('./client').Client;
api.build = function (opts) {
  var c = new api.Client(opts);

  _.each(api.modules, function (module, name) {
    c.attach(name, module);
  });

  return c;
};
