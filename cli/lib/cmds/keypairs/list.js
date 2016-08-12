'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var keypairs = require('../../keypairs');

var cmd = new Command(
  'keypairs',
  'List keypairs for an organization.',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      keypairs.list.execute(ctx).then(function (payload) {
        keypairs.list.output.success(null, ctx, payload);

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

flags.add(cmd, 'org');
flags.add(cmd, 'all', {
  description: 'if given, list keys for all organizations'
});

module.exports = cmd;
