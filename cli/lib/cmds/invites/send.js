'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var invites = require('../../invites');

var cmd = new Command(
  'invites:send <email>',
  'send an invitation to an email address to join an organization',
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
  '-t, --team <name>', 'a team the user should be added too', undefined);

module.exports = cmd;
