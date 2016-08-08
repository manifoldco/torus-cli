'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');

var orgs = require('../../orgs');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'orgs:create [name]',
  'Create an organization',
  'knotty-buoy\n\n  Creates a new organization called \'knotty-buoy\'',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      orgs.create.execute(ctx).then(function () {
        orgs.create.output.success();
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        orgs.create.output.failure(err);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

module.exports = cmd;
