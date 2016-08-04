'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var invites = require('../../invites');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'invites:approve <email>',
  'approve an invitation that was sent to an email address to join the org',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      invites.approve.execute(ctx).then(function (payload) {
        invites.approve.output.success(null, payload);

        resolve(true);
      }).catch(function (err) {
        invites.approve.output.failure();

        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');

module.exports = cmd;
