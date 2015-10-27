'use strict';
var log = require('../util/log').get('help');

module.exports.command = function help (argv, cmds) {
  log.print('Arigato amazing CLI tool');
};
