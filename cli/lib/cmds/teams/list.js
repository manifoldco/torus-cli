'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var list = require('../../teams/list');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'teams',
  'List teams in an organization',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      list.execute(ctx).then(function (teams) {
        list.output.success(null, teams);
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        list.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');

module.exports = cmd;
