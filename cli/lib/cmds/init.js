'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../cli/command');

var init = require('../context/init');
var auth = require('../middleware/auth');

var cmd = new Command(
  'init',
  'initalize a project and service linking to a codebase',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      init.execute(ctx).then(function (result) {
        init.output.success(result);

        resolve(true);
      }).catch(function (err) {
        err.type = err.type || 'unknown';

        init.output.failure();
        reject(err);
      });
    });
  }
);

cmd.hook('pre', auth());

module.exports = cmd;
