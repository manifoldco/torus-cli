'use strict';

var Promise = require('es6-promise').Promise;

var targetMap = require('./map');

var unlink = exports;
unlink.output = {};

unlink.output.success = function (ctx, results) {
  var target = results.target;

  // target.path is guaranteed to be a substring of process.cwd()
  var isParent = (process.cwd().length === target.path.length);
  if (isParent) {
    return console.log('Your current working directory has been unlinked.');
  }

  return console.log(
    'The parent directory (' + target.path + ') has been unlinked.');
};

unlink.output.failure = function () {
  console.log('Failed to unlink your current working directory.');
};

unlink.execute = function (ctx) {
  return new Promise(function (resolve, reject) {
    return targetMap.derive(ctx.config, process.cwd()).then(function (targets) {
      if (targets.length === 0) {
        throw new Error('You must be in a linked directory.');
      }

      var targetToRemove = targets[0];

      return targetMap.unlink(ctx.config, targetToRemove).then(function () {
        resolve({
          target: targetToRemove
        });
      });
    }).catch(reject);
  });
};
