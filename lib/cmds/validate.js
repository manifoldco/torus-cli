'use strict';

const log = require('../util/log').get('cmds/validate');
const CommandInterface = require('../command').CommandInterface;

const ArigatoConfigError = require('../errors').ArigatoConfigError;
const Arigato = require('../descriptors/arigato').Arigato;
const ARIGATO_FILE_NAME = require('../descriptors/arigato').FILE_NAME;

class Validate extends CommandInterface {

  execute () {
    return new Promise((resolve, reject) => {
      Arigato.find(process.cwd()).then((arigato) => {
        log.print(`${arigato.path} is valid!`);
        resolve(true);
      }).catch((errors) => {
        if (!Array.isArray(errors) && !(errors instanceof ArigatoConfigError)) {
          return reject(errors);
        }

        errors = (Array.isArray(errors)) ? errors : [errors];
        log.print(`Found ${errors.length} errors\n`);

        errors.forEach((err) => {
          log.print(`\t${err.message}`);
        });

        resolve(false);
      });
    });
  }
}

module.exports = {
  key: 'validate',
  synopsis: 'validate an '+ARIGATO_FILE_NAME+' for errors',
  usage: 'arigato validate',
  Command: Validate,
  example: `\tlocalhost$ arigato validate`
};
