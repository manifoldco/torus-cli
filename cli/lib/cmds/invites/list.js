'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var invites = require('../../invites');

var cmd = new Command(
  'invites',
  'List outstanding invitations for an organization. These invites have yet to be approved',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      invites.list.execute(ctx).then(function (payload) {
        invites.list.output.success(null, ctx, payload);

        resolve(true);
      }).catch(function (err) {
        invites.list.output.failure();

        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');

cmd.option(
  '-a, --approved', 'only list invites which have been confirmed', false);

module.exports = cmd;
