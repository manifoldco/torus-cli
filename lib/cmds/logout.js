'use strict';

const CommandInterface = require('../command').CommandInterface;
const users = require('../users');
const errors = require('../errors');

const log = require('../util/log').get('cmds/logout');

class Logout extends CommandInterface {
  execute () {
    return users.logout().then(() => {
      log.print('You\'ve successfully logged out!');
      return Promise.resolve(true);
    }).catch((err) => {
      if (err instanceof errors.NotAuthenticatedError) {
        log.error('You cannot log out if you\'re not authenticated!');
        return Promise.resolve(false);
      }

      return Promise.reject(err);
    }); 
  }
}

module.exports = {
  key: 'logout',
  synopsis: 'logs the current user out of their session',
  usage: 'arigato logout',
  Command: Logout,
  example: `\tlocalhost$ arigato logout`
};
