'use strict';

var Promise = require('es6-promise').Promise;

var flags = require('../../flags');
var Command = require('../../cli/command');
var create = require('../../projects/create');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var example = 'landing-page --org knotty-buoy\n\n';
example += '  Creates a project called \'landing-page\' in the organization \'knott-buoy\'';

var cmd = new Command(
  'projects:create [name]',
  'Create a project in an organization',
  example,
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
  description: 'Organization this project will belong to'
});

module.exports = cmd;
