'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var flags = require('../flags');
var link = require('../context/link');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var cmd = new Command(
  'link',
  'Link your current directory to Arigato',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      link.execute(ctx).then(function (result) {
        link.output.success(ctx, result);

        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';

        link.output.failure(ctx);
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

flags.add(cmd, 'force', {
  description: 'Overwrite the organization and project currently linked'
});

module.exports = cmd;
