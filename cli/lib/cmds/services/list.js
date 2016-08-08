'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var services = require('../../services');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'services',
  'List services for an organization',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      services.list.execute(ctx).then(function (results) {
        services.list.output.success(null, results);

        resolve(true);
      }).catch(function (err) {
        services.list.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');
flags.add(cmd, 'project');

module.exports = cmd;
