'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var invites = require('../../invites');

var cmd = new Command(
  'invites:send <email>',
  'Send an invitation to join an organization to an email address',
  'jeff@example.com --team api-ops\n\n  Sends an invite to Jeff to join the api-ops team',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      invites.send.execute(ctx).then(function (payload) {
        invites.send.output.success(null, payload);

        resolve(true);
      }).catch(function (err) {
        invites.send.output.failure();

        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');

cmd.option(
  '-t, --team <name>', 'Team to which the user will be added', undefined);

module.exports = cmd;
