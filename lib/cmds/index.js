'use strict';
const fs = require('fs');

const cbwrap = require('cbwrap');

const log = require('../util/log').get('main');
const readdir = cbwrap.wrap(fs.readdir);

const COMMAND_REGISTRY = {};
const JS_REGEX = /^.*\.js$/;
const REPLACE_EXT = /\.js$/;

let cmds = exports;

cmds.load = function () {
  return new Promise((resolve, reject) => {
    readdir(__dirname).then((files) => {
      files = files.filter((filename) => {
        return (filename.match(JS_REGEX) && filename !== 'index.js');
      });

      const names = files.map((filename) => filename.replace(REPLACE_EXT, ''));

      names.forEach((name) => {
        var cmdModule = require(`./${name}`);
        COMMAND_REGISTRY[cmdModule.command] = cmdModule;
      });
     
      resolve();
    }).catch(reject);
  });
};

// TODO: Write the help function dawg
cmds.help = function () {
  return Promise.resolve(); 
};

cmds.process = function (argv) {

  if (argv._.length === 0) {
    return cmds.help();
  }

  const input = argv._[0].toLowerCase();

  if (input === 'help') {
    return cmds.help();
  }

  // TODO: Figure out a better command router mechanism
  const cmdModule = COMMAND_REGISTRY[input.toLowerCase()];
  if (!cmdModule) {
    log.error('Unknown command: ', cmdModule);
    return cmds.help();
  }

  // TODO: Come back and make this 'cancellabe'.. meaning it handles process
  // signals and other things gracefully.
  const cmd = new cmdModule.Command(argv);
  return cmd.execute();
};
