'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var link = require('../context/link');
var auth = require('../middleware/auth');

var cmd = new Command(
  'link',
  'setup a link between this working directory and the arigato cloud',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      link.execute(ctx).then(function (result) {
        link.output.success(ctx, result);

        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';

        link.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());

module.exports = cmd;
