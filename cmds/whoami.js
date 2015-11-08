'use strict';
var log = require('../util/log').get('whoami');

module.exports.command = function whoami (argv, cmds) {
  log.print('I have no fucking idea bro :)');
};

module.exports.helpText = `
Hey this is some help text!
`;
