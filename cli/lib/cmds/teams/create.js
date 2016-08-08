'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../../flags');
var Command = require('../../cli/command');
var create = require('../../teams/create');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'teams:create [name]',
  'Create and add a new team to an organization',
  'api-ops --org knotty-buoy\n\n Creates a new team called \'api-ops\' and adds it to the org',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      create.execute(ctx).then(function () {
        create.output.success();

        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        create.output.failure(err);
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
