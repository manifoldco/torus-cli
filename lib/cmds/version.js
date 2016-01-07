'use strict';

const pkg = require('../../package.json');
const log = require('../util/log').get('version');
const CommandInterface = require('../command').CommandInterface;

const VERSION = pkg.version;

class Version extends CommandInterface {
  execute () {
    log.print(VERSION);
    return Promise.resolve(true); 
  }
}

module.exports = {
  key: 'version',
  synopsis: 'prints the current version of the Arigato CLI',
  Command: Version
};
