'use strict';
var log = require('../util/log').get('whoami');

module.exports.command = function whoami (argv, cmds) {
  log.print('I have no fucking idea bro :)');
};
