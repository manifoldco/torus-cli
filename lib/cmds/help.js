'use strict';

const CommandInterface = require('../command').CommandInterface;

class Help extends CommandInterface {}

module.exports = {
  key: 'help',
  synopsis: 'lists details about specific commands',
  usage: 'arigato help',
  Command: Help,
  example: '\tlocalhost$ arigato help'
};
