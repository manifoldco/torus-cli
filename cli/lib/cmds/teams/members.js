'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../../flags');
var Command = require('../../cli/command');
var members = require('../../teams/members');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'teams:members [name]',
  'List the members of a team in an organization',
  'api-ops --org knotty-buoy\n\n Shows all members of a team called \'api-ops\'',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      members.execute(ctx).then(function (data) {
        members.output.success(null, data);
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        members.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org', {
  description: 'Organization this team will belong to'
});

module.exports = cmd;
