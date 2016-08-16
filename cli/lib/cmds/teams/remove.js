'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../../flags');
var Command = require('../../cli/command');
var remove = require('../../teams/remove');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'teams:remove [username] [team]',
  'Remove user from a specified team in an organization you administer',
  'jeff api-ops --org knotty-buoy\n\n  Removes Jeff from the api-ops team',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      remove.execute(ctx).then(function (payload) {
        remove.output.success(null, payload);

        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        remove.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org', {
  description: 'Organization of the team you wish to remove a member from'
});

module.exports = cmd;
