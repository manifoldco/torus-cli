'use strict';

const version = require('../util/version');
const log = require('../util/log').get('version');
const CommandInterface = require('../command').CommandInterface;


class Version extends CommandInterface {
  execute () {
    log.print(version.get());
    return Promise.resolve(true); 
  }
}

module.exports = {
  key: 'version',
  synopsis: 'prints the current version of the Arigato CLI',
  usage: 'arigato version',
  Command: Version,
  example: '\tlocalhost$ arigato version'
};

