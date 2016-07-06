'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../../flags');
var Command = require('../../cli/command');
var create = require('../../teams/create');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'teams:create [name]',
  'add a new team to an organization',
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
  description: 'the organization this team will belong to'
});

module.exports = cmd;
