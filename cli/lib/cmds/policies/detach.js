'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var flags = require('../../flags');
var auth = require('../../middleware/auth');
var target = require('../../middleware/target');
var detach = require('../../policies/detach');

var example = 'basic api-ops --org knotty-buoy\n\n  ';
example += 'Removes the policy \'basic\' from the api-ops team in the org \'knotty-buoy\'';

var cmd = new Command(
  'policies:detach [name] [team]',
  'Detach a policy from a team, does not delete the policy',
  example,
  function (ctx) {
    return new Promise(function (resolve, reject) {
      detach.execute(ctx).then(function (payload) {
        detach.output.success(null, payload);

        resolve(true);
      }).catch(function (err) {
        detach.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'org');

module.exports = cmd;
