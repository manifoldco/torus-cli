/**
 * Utility library for standardizing flags, their naming, and the defaults.
 */

'use strict';

var Command = require('./cli/command');

var flags = exports;

var LIST = flags.LIST = {
  all: {
    usage: '-a, --all',
    description: 'Perform command on all organizations',
    default: false
  },
  force: {
    usage: '-f, --force',
    description: 'Force override',
    default: undefined
  },
  project: {
    usage: '-p, --project [name]',
    description: 'Specify the project',
    default: undefined
  },
  org: {
    usage: '-o, --org [name]',
    description: 'Specify the organization',
    default: undefined
  },
  service: {
    usage: '-s, --service [name]',
    description: 'Specify the service',
    default: undefined,
    allowFromEnv: true
  },
  environment: {
    usage: '-e, --environment [name]',
    description: 'Specify the environment',
    default: undefined,
    allowFromEnv: true
  },
  user: {
    usage: '-u, --user [username]',
    description: 'Specify the user'
  },
  instance: {
    usage: '-i, --instance [name]',
    description: 'Specify the instance',
    default: '1'
  }
};

var ENV_VAR_PREFIX = 'AG_';

flags.add = function (cmd, name, overrides) {
  if (!(cmd instanceof Command)) {
    throw new Error('Must provide an instance of a Command');
  }

  if (!LIST[name]) {
    throw new Error('Unknown option: ' + name);
  }

  var matching = cmd.options.filter(function (o) {
    return (o.name() === name);
  });

  if (matching.length > 0) {
    throw new Error('Cannot add the same option twice');
  }

  overrides = overrides || {};

  var opt = LIST[name];

  var defaultValue = overrides.default || opt.default;
  var envVar = process.env[ENV_VAR_PREFIX + name.toUpperCase()];
  if (opt.allowFromEnv === true) {
    defaultValue = envVar;
  }

  cmd.option.apply(cmd, [
    opt.usage,
    (overrides.description) ? overrides.description : opt.description,
    defaultValue
  ]);
};
