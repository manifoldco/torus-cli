'use strict';

const log = require('../util/log').get('cmds/init');
const CommandInterface = require('../command').CommandInterface;

const errors = require('../errors');
const users = require('../users');
const apps = require('../apps');

class Init extends CommandInterface {

  execute (argv) {
    return new Promise((resolve, reject) => {
      var appName = argv._[0];
      users.loggedIn().then((loggedIn) => {
        if (!loggedIn) {
          log.error('You must be logged in to initialize an app.');
          return resolve(false);
        }

        return apps.init(appName);
      }).then(() => {
        log.print('Your application has been created');
        log.print('An arigato.yaml file has been added to your project root');
        resolve(true);
      }).catch((err) => {
        if (err instanceof errors.ValidationError) {
          log.error('Invalid app name provided: '+err.message);
          return resolve(false);
        }

        reject(err);
      });
    });
  }
}

module.exports = {
  key: 'init',
  synopsis: 'initialzie an arigato app in the root of your project',
  usage: 'arigato init <my-app>',
  Command: Init,
  example: `\tlocalhost$ arigato init api`
};
