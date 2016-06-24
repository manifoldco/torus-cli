'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../flags');
var Command = require('../cli/command');
var run = require('../credentials/run');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var runCmd = new Command(
  'run <command>',
  'run a process and inject credentials into its environment',
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
  description: 'the org the credentials belong to'
});
flags.add(runCmd, 'project', {
  description: 'the project the credentials belong to'
});
flags.add(runCmd, 'service', {
  description: 'the service the credentials belong to'
});
flags.add(runCmd, 'environment', {
  description: 'the environment the credentiasl belong to'
});
flags.add(runCmd, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

module.exports = runCmd;
