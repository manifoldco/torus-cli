'use strict';
var log = require('../util/log').get('whoami');

let CommandInterface = require('../command').CommandInterface;

class WhoAmI extends CommandInterface {
  execute () {
    return new Promise((resolve) => {
      log.print(`hello im not sure who you are yet!`);
      resolve(true);
    });
  }
}

module.exports = {
  key: 'whoami',
  synopsis: 'prints information about the current logged in user',
  Command: WhoAmI 
};
