'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var list = require('../../policies/list');

var cmd = new Command(
  'policies',
  'list policies associated with an organization',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      list.execute(ctx).then(function (payload) {
        list.output.success(null, payload);

        resolve(true);
      }).catch(function (err) {
        list.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');

module.exports = cmd;
