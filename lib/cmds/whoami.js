'use strict';
var log = require('../util/log').get('whoami');

const CommandInterface = require('../command').CommandInterface;
const users = require('../users');

class WhoAmI extends CommandInterface {
  execute () {
    return new Promise((resolve, reject) => {
      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('I\'m not sure who you are, perhaps you should login?');
          return resolve(false);
        }

        return users.me();
      }).then((user) => {
        log.print('You\'re logged in as '+user.name);
      }).catch(reject);
    });
  }
}

module.exports = {
  key: 'whoami',
  synopsis: 'prints information about the current logged in user',
  Command: WhoAmI
};
