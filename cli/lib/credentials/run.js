'use strict';

var childProcess = require('child_process');

var _ = require('lodash');
var Promise = require('es6-promise').Promise;
var regulator = require('event-regulator');

var cValue = require('./value');
var credentials = require('./credentials');
var harvest = require('./harvest');

var run = exports;

var ENV_BLACKLIST = [
  'ag_email',
  'ag_password',
  'ag_environment',
  'ag_service'
];

run.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    var params = harvest.get(ctx);

    return credentials.get(ctx.api, params).then(function (results) {
      var creds = results.credentials;

      return run.spawn(ctx.params, creds).then(resolve);
    }).catch(reject);
  });
};

run.spawn = function (params, creds) {
  return new Promise(function (resolve, reject) {
    if (params.length === 0) {
      throw new Error('Missing command parameter');
    }

    var env = _.clone(process.env);
    creds.forEach(function (cred) {
      var value = cValue.parse(cred.body.value);
      if (value.body.type === 'undefined') {
        return;
      }

      env[cred.body.name.toUpperCase()] = value.body.value;
    });

    // Remove blacklisted environment variables, do not pass to child process
    _.each(env, function (v, k) {
      if (ENV_BLACKLIST.indexOf(k.toLowerCase()) > -1) {
        delete env[k];
      }
    });

    var proc = childProcess.spawn(params[0], params.slice(1), {
      cwd: process.cwd(),
      detached: false,
      env: env
    });

    proc.stdout.pipe(process.stdout);
    proc.stderr.pipe(process.stderr);

    function handleSignal(signal) {
      proc.kill(signal);
    }

    regulator.subscribe([
      [process, 'SIGHUP', handleSignal.bind(this, 'SIGHUP')],
      [process, 'SIGINT', handleSignal.bind(this, 'SIGINT')],
      [process, 'SIGQUIT', handleSignal.bind(this, 'SIGQUIT')],
      [process, 'SIGTERM', handleSignal.bind(this, 'SIGTERM')],

      [proc, 'error', reject, true],
      [proc, 'close', resolve, true]
    ]);
  });
};
