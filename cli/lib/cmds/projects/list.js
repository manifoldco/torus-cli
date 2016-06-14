'use strict';

var Promise = require('es6-promise').Promise;

var Command = require('../../cli/command');
var list = require('../../projects/list');
var auth = require('../../middleware/auth');

var cmd = new Command(
  'projects',
  'list all projects in an organization',
  function (ctx) {
    return new Promise(function (resolve, reject) {
      list.execute(ctx).then(function (projects) {
        list.output.success(null, projects);
        resolve();
      }).catch(function (err) {
        err.type = err.type || 'unknown';
        list.output.failure(err);
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
