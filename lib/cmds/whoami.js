'use strict';
var log = require('../util/log').get('whoami');

let CommandInterface = require('../command').CommandInterface;

class WhoAmI extends CommandInterface {
  constructor () {
    super();
  }

  execute () {
    return new Promise((resolve) => {
      log.print(`hello I\'m not sure who you are yet!`);
      resolve();
    });
  }
}

module.exports = {
  command: 'whoami',
  Command: WhoAmI 
};
