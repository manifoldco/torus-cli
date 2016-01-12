'use strict';
const fs = require('fs');

const _ = require('lodash');
const cbwrap = require('cbwrap');

const pad = require('../util/pad');
const log = require('../util/log').get('main');
const readdir = cbwrap.wrap(fs.readdir);

let cmds = exports;

const COMMAND_REGISTRY = cmds.COMMAND_REGISTRY = {};
const JS_REGEX = /^.*\.js$/;
const REPLACE_EXT = /\.js$/;

cmds.load = function () {
  return new Promise((resolve, reject) => {
    readdir(__dirname).then((files) => {
      files = files.filter((filename) => {
        return (filename.match(JS_REGEX) && filename !== 'index.js');
      });

      const names = files.map((filename) => filename.replace(REPLACE_EXT, ''));

      names.forEach((name) => {
        var cmdModule = require(`./${name}`);
        COMMAND_REGISTRY[cmdModule.key] = cmdModule;
      });

      resolve();
    }).catch(reject);
  });
};

// TODO: Abstract the "route" functionality based on the order of pure values
// given to us from yargs argv._ -- ideally, we can re-use this throughout
//
// XXX: Perhaps this should iterate down argv._ and finding help topics.
// It'd then print them.. it'd be easier to keep things in sync then.
cmds.help = function (argv) {

  if (!argv || argv._.length === 0) {
    return cmds._mainHelp();
  }

  const input = argv._[0].toLowerCase();
  const cmdModule = cmds._getModule(input);
  if (!cmdModule) {
    log.error('Unknown help topic:', input);
    return Promise.resolve(false);
  }

  return cmds._printHelp(cmdModule);
};

cmds.process = function (argv) {
  if (argv._.length === 0) {
    return cmds.help();
  }

  const input = argv._[0].toLowerCase();
  argv._.shift();
  if (input === 'help') {
    return cmds.help(argv);
  }

  // TODO: Figure out a better command router mechanism
  const cmdModule = cmds._getModule(input);
  if (!cmdModule) {
    log.error('Unknown command:', input);
    return Promise.resolve(false);
  }

  // TODO: Come back and make this 'cancellabe'.. meaning it handles process
  // signals and other things gracefully.
  const cmd = new cmdModule.Command();
  return cmd.execute(argv);
};

cmds._mainHelp = function () {
  const usage = `Usage: arigato [-OPTS] [COMMAND] [PARAMETERS...]`;
  let help =
    `To view additional details for a command, type "arigato help [COMMAND]"`;

  let cmdList = '';
  _.each(cmds.COMMAND_REGISTRY, function (mod) {
    let cmdPadded = pad.right(mod.key, 13);
    cmdList += `    ${cmdPadded}  ${mod.synopsis}\n`;
  });

  log.print(`${usage}\n\n${help}\n\n${cmdList}`);
  return Promise.resolve(false);
};

cmds._printHelp = function (cmdModule) {

  const usage = `Usage: ${cmdModule.usage}\n\n${cmdModule.synopsis}`;
  log.print(`${usage}\n\nExample(s):\n\n${cmdModule.example}`);

  return Promise.resolve(false);
};

cmds._getModule = function (name) {
  return COMMAND_REGISTRY[name.toLowerCase()];
};
