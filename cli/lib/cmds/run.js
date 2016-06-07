'use strict';

var Promise = require('es6-promise').Promise;

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

runCmd.option(
  '-s, --service [service]',
  'the service this process belongs too',
  undefined
);

runCmd.option(
  '-e, --environment [environment]',
  'the environment this process is running in',
  undefined
);

runCmd.option(
  '-i, --instance [name]',
  'the instance of the service',
  '1'
);

module.exports = runCmd;
