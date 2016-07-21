'use strict';

var path = require('path');
var fs = require('fs');
var Promise = require('es6-promise').Promise;

var pkg = require('../package.json');
var Program = require('./cli/program');
var cmds = require('./cmds');

var config = require('./middleware/config');
var prefs = require('./middleware/prefs');
var daemon = require('./middleware/daemon');
var api = require('./middleware/api');

var arigato = exports;

arigato.run = function (opts) {
  return new Promise(function (resolve, reject) {
    var templates = {
      program: fs.readFileSync(
        path.join(__dirname, '../templates/program.template')),
      command: fs.readFileSync(
        path.join(__dirname, '../templates/command.template'))
    };

    var program = new Program('arigato', pkg.version, templates);
    program.hook('pre', config(opts.arigatoRoot));
    program.hook('pre', prefs());
    program.hook('pre', daemon.preHook());
    program.hook('pre', api());

    cmds.get().then(function (cmdList) {
      cmdList.forEach(program.command.bind(program));

      return program.run(opts.argv).then(resolve);
    }).catch(reject);
  });
};
