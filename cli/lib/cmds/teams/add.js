'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../../flags');
var Command = require('../../cli/command');
var add = require('../../teams/add');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'teams:add [username] [team]',
  'add user to a team in an org you administor',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      add.execute(ctx).then(function (payload) {
        add.output.success(null, payload);

        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        add.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org', {
  description: 'the organization of the team you wish to add a member to'
});

module.exports = cmd;
