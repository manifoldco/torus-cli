'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var keypairs = require('../../keypairs');

var cmd = new Command(
  'keypairs:generate',
  'Generate a new keypair for an organization',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      keypairs.generate.execute(ctx).then(function () {
        resolve(true);
      }).catch(function (err) {
        keypairs.list.output.failure();

        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'all', {
  description: 'if given, keys will be generated for all orgs without valid keys'
});
flags.add(cmd, 'org');

module.exports = cmd;
