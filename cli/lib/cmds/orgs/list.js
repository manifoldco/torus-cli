'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var orgs = require('../../orgs');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');

var cmd = new Command(
  'orgs',
  'List organizations associated with your account',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      orgs.list.execute(ctx).then(function (payload) {
        orgs.list.output.success(null, payload);

        resolve(true);
      }).catch(function (err) {
        orgs.list.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

module.exports = cmd;
