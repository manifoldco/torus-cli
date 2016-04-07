'use strict';

var path = require('path');
var fs = require('fs');
var Promise = require('es6-promise').Promise;

var pkg = require('../../package.json');
var Program = require('./cli/program');
var cmds = require('./cmds');

var arigato = exports;

arigato.run = function (argv) {
  return new Promise(function (resolve, reject) {
    var templates = {
      program: fs.readFileSync(
        path.join(__dirname, '../templates/program.template')),
      command: fs.readFileSync(
        path.join(__dirname, '../templates/command.template'))
    };

    var program = new Program('arigato', pkg.version, templates);
    cmds.get().then(function(cmdList) { 
      cmdList.forEach(program.command.bind(program));

      program.run(argv).then(resolve).catch(reject);
    }).catch(reject);
  });
};
