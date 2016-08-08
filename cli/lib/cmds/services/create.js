'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var flags = require('../../flags');
var services = require('../../services');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'services:create [name]',
  'Create a service in a project in an organization',
  'www --org knotty-buoy\n\n  Creates a new service called \'www\' in the org \'knotty-buoy\'',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      services.create.execute(ctx).then(function () {
        services.create.output.success();
        resolve();

      // Service creation failed
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        services.create.output.failure(err);
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
