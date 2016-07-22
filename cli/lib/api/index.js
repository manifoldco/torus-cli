/* eslint global-require: "off" */
'use strict';

var _ = require('lodash');

var api = exports;

api.modules = {
  users: require('./users'),
  orgs: require('./orgs'),
  teams: require('./teams'),
  policies: require('./policies'),
  policyAttachments: require('./policy_attachments'),
  memberships: require('./memberships'),
  projects: require('./projects'),
  envs: require('./envs'),
  services: require('./services'),
  credentials: require('./credentials'),
  versionApi: require('./versionApi'),
  login: require('./login'),
  logout: require('./logout'),
  session: require('./session')
};

api.Client = require('./client').Client;
api.build = function (opts) {
  var c = new api.Client(opts);

  _.each(api.modules, function (module, name) {
    c.attach(name, module);
  });

  return c;
};
