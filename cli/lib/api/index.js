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
  invites: require('./invites'),
  profiles: require('./profiles'),

  // non-proxied daemon apis
  versionApi: require('./versionApi'),
  login: require('./login'),
  session: require('./session'),
  keypairs: require('./keypairs')
};

api.Client = require('./client').Client;
api.build = function (opts) {
  var c = new api.Client(opts);

  _.each(api.modules, function (module, name) {
    c.attach(name, module);
  });

  return c;
};
