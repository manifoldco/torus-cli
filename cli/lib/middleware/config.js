'use strict';

var fs = require('fs');
var path = require('path');

var Promise = require('es6-promise').Promise;

var config = {};
module.exports = config;

var FOLDER_PERM_STRING = '0700';
var FOLDER_PERM = 0o700;

config.middleware = function (arigatoRoot) {

  arigatoRoot = arigatoRoot || path.join(process.env.HOME, '.arigato');

  return function (ctx) {
    return new Promise(function (resolve, reject) {

      function createFolder() {
        fs.mkdir(arigatoRoot, FOLDER_PERM, function (err) {
          if (err) {
            return reject(err);
          }

          resolve();
        });
      }
      ctx.config = new Config(arigatoRoot, ctx.program.version);

      fs.stat(arigatoRoot, function (err, stat) {
        if (err && err.code === 'ENOENT') {
          return createFolder();
        }
        if (err) {
          return reject(err);
        }

        if (!stat.isDirectory()) {
          return reject(
            new Error('Arigato Root must be a directory: '+arigatoRoot));
        }
    
        var fileMode = '0' + (stat.mode & parseInt('777',8)).toString(8);
        if (fileMode !== FOLDER_PERM_STRING) {
          return reject(new Error(
            'Arigato root file permission error: '+arigatoRoot+' '+fileMode+
            ' not ' + FOLDER_PERM_STRING
          ));
        }

        resolve();
      });
    });
  };
};

function Config (arigatoRoot, version) {
  this.arigatoRoot = arigatoRoot;
  this.socketPath = path.join(arigatoRoot, 'daemon.socket');
  this.pidPath = path.join(arigatoRoot, 'daemon.pid');
  this.version = version;
}
config.Config = Config;
