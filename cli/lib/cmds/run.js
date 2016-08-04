'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../flags');
var Command = require('../cli/command');
var run = require('../credentials/run');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var runCmd = new Command(
  'run <command>',
  'run a process and inject secrets into its environment',
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
  description: 'the org the secrets belong to'
});
flags.add(runCmd, 'project', {
  description: 'the project the secrets belong to'
});
flags.add(runCmd, 'service', {
  description: 'the service the secrets belong to'
});
flags.add(runCmd, 'environment', {
  description: 'the environment the secrets belong to'
});
flags.add(runCmd, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

module.exports = runCmd;
