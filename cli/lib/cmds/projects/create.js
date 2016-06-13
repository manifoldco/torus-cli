'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var create = require('../../projects/create');
var auth = require('../../middleware/auth');

var cmd = new Command(
  'projects:create [name]',
  'create a project under an organization',
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

cmd.option(
  '-o, --org [name]',
  'the organization this project will belong too',
  undefined
);

module.exports = cmd;
