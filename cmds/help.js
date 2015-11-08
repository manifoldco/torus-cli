'use strict';
var log = require('../util/log').get('help');

module.exports.command = function help (argv, cmds) {
  
  if (argv._.length === 0) {
    return log.print('Arigato amazing CLI tool');
  }

  var target = cmds[argv._[0]];
  if (target && target.helpText) {
    return log.print(target.helpText);
  }

  log.error('Unknown Help Command: ', target);
  return 1;
};
