'use strict';

var Promise = require('es6-promise').Promise;

var targetMap = require('./map');

var unlink = exports;
unlink.output = {};

unlink.output.success = function (ctx) {
  var projectMap = ctx.target ? ctx.target.path() : '';
  var currentPath = targetMap.path();
  var isParent = (currentPath.length === projectMap.length);

  if (isParent) {
    return console.log('\nYour current working directory has been unlinked.\n');
  }

  return console.log(
    '\nThe parent directory (' + projectMap + ') has been unlinked.\n');
};

unlink.output.failure = function () {
  console.log('\nFailed to unlink your current working directory.\n');
};

unlink.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    return targetMap.unlink(ctx.target).then(function () {
      return ctx.target;
    })
    .then(resolve)
    .catch(reject);
  });
};
