'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../flags');
var Command = require('../cli/command');
var invite = require('../user/invite');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var inviteCmd = new Command(
  'invite <username>',
  'invite a user to collaborate in your org',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      invite.execute(ctx).then(function (username) {
        invite.output.success(null, username);
        resolve();
      }).catch(function (err) {
        invite.output.failure();
        reject(err);
      });
    });
  }
);

inviteCmd.hook('pre', auth());
inviteCmd.hook('pre', target());

flags.add(inviteCmd, 'org', {
  description: 'the org to invite the user to'
});

module.exports = inviteCmd;
