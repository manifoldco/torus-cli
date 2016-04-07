'use strict';

var path = require('path');
var glob = require('glob');
var Promise = require('es6-promise').Promise;

var cmds = exports;

cmds._cache = null;
cmds.get = function () {

  if (cmds._cache) {
    return cmds._cache;
  }

  var p = new Promise(function (resolve, reject) {
    glob(path.join(__dirname+'/**/*.js'), function (err, files) {
      if (err) {
        return reject(err);
      }

      files = files.filter(function(name) {
        return (name.indexOf('index.js') === -1);
      });

      var objs = files.map(function(name) { 
        return require(name);
      });

      resolve(objs);
    });
  });

  cmds._cache = p;
  return p;
};
