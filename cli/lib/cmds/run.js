'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../flags');
var Command = require('../cli/command');
var run = require('../credentials/run');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var runCmd = new Command(
  'run <command>',
  'Run a process and inject secrets into its environment',
  './bin/www\n\n  Runs the www process with Arigato injecting secrets',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      run.execute(ctx).then(function (exitCode) {
        resolve(exitCode);
      }).catch(function (err) {
        reject(err);
      });
    });
  }
);

runCmd.hook('pre', auth());
runCmd.hook('pre', target());

flags.add(runCmd, 'org', {
  description: 'Organization the secrets belong to'
});
flags.add(runCmd, 'project', {
  description: 'Project the secrets belong to'
});
flags.add(runCmd, 'service', {
  description: 'Service the secrets belong to'
});
flags.add(runCmd, 'environment', {
  description: 'Environment the secrets belong to'
});
flags.add(runCmd, 'instance', {
  description: 'Instance of the service belonging to the current user'
});

module.exports = runCmd;
