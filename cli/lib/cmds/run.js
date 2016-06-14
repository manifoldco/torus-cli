'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../flags');
var Command = require('../cli/command');
var run = require('../credentials/run');
var auth = require('../middleware/auth');

var runCmd = new Command(
  'run <command>',
  'run a process and inject credentials into its environment',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      run.execute(ctx).then(function (success) {
        resolve(success);
      }).catch(function (err) {
        reject(err);
      });
    });
  }
);

runCmd.hook('pre', auth());

flags.add(runCmd, 'org', {
  description: 'the org the credentials belongs too'
});
flags.add(runCmd, 'service', {
  description: 'the service the credentials belong too'
});
flags.add(runCmd, 'environment', {
  description: 'the environment the credentiasl belong too'
});
flags.add(runCmd, 'instance', {
  description: 'the instance of the service belonging to the current user'
});

module.exports = runCmd;
