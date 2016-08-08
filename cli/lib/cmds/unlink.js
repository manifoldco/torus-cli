'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var unlink = require('../context/unlink');
var auth = require('../middleware/auth');
var target = require('../middleware/target');

var cmd = new Command(
  'unlink',
  'Remove the link between this project and Arigato',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      unlink.execute(ctx).then(function (result) {
        unlink.output.success(ctx, result);

        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';

        unlink.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());
cmd.hook('pre', target());

module.exports = cmd;
