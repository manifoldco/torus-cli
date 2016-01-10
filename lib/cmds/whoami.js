'use strict';
var log = require('../util/log').get('whoami');

let CommandInterface = require('../command').CommandInterface;
let users = require('../users');

class WhoAmI extends CommandInterface {
  execute () {
    return new Promise((resolve) => {
      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('I\'m not sure who you are, perhaps you should login?');
          return resolve(false);
        }

        log.print('You are someone I use to know.');
        resolve(true);
      });
    });
  }
}

module.exports = {
  key: 'whoami',
  synopsis: 'prints information about the current logged in user',
  Command: WhoAmI 
};
